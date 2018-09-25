package phly

import (
	"errors"
	"fmt"
	"github.com/micro-go/parse"
	"io"
	"os"
	"strings"
	"sync"
)

// --------------------------------
// PIPELINE interface

type Pipeline interface {
	Start(StartArgs) error
	Stop() error
	Wait() error
}

// --------------------------------
// PIPELINE start

// StartArgs provides arguments when starting the pipeline.
type StartArgs struct {
	Cla map[string]string // Command line arguments
}

// --------------------------------
// PIPELINE struct

type pipeline struct {
	workingdir  string                `json:"-"`
	args        pipeline_args         `json:"-"`
	file        string                `json:"-"`
	ins         []connection          `json:"-"`
	nodes       map[string]*container `json:"-"`
	inputDescr  []pipelinePinDescr    `json:"-"`
	outputDescr []pipelinePinDescr    `json:"-"`
	// running
	stop chan struct{}  `json:"-"`
	wg   sync.WaitGroup `json:"-"`
	run  *runner        `json:"-"`
}

func (p *pipeline) Describe() NodeDescr {
	descr := NodeDescr{Id: "phly/pipeline", Name: "Pipeline", Purpose: "Run an internal pipeline."}
	for _, pin := range p.inputDescr {
		descr.InputPins = append(descr.InputPins, PinDescr{Name: pin.Name, Purpose: pin.Purpose})
	}
	for _, pin := range p.outputDescr {
		descr.OutputPins = append(descr.OutputPins, PinDescr{Name: pin.Name, Purpose: pin.Purpose})
	}
	return descr
}

func (p *pipeline) Instantiate(args InstantiateArgs, cfg interface{}) (Node, error) {
	ans := &pipeline{}
	file, _ := parse.FindTreeString("file", cfg)
	if file != "" {
		ans.file = file
		r := args.Env.FindReader(file)
		if r == nil {
			return nil, errors.New("Pipeline can't find " + file)
		}
		err := readPipeline(r, ans)
		if err != nil {
			return nil, err
		}
	}
	return ans, nil
}

func (p *pipeline) Run(args RunArgs, input Pins, sender PinSender) (Flow, error) {
	// Override the run args with my values
	if args.stop == nil {
		args.stop = make(chan struct{})
	}
	// Clear out any previous run
	p.stop = args.stop
	p.Wait()

	args.workingdir = p.workingdir

	inputs, err := p.gatherInputs(args, input)
	if err != nil {
		return nil, err
	}
	sources, err := p.gatherSources(inputs)
	if err != nil {
		return nil, err
	}
	r := &runner{}
	_, err = r.runAsync(args, p, sources, inputs)
	if err == nil {
		p.run = r
	}
	return Running, err
}

func (p *pipeline) Start(sa StartArgs) error {
	p.Stop()
	input := &pins{}
	args := RunArgs{Env: env, cla: sa.Cla, DryRun: dryrun}
	_, err := p.Run(args, input, nil)
	return err
}

func (p *pipeline) Stop() error {
	if p.stop != nil {
		close(p.stop)
		p.stop = nil
	}
	return nil
}

func (p *pipeline) Wait() error {
	p.wg.Wait()
	var err error
	if p.run != nil {
		err = p.run.Error()
		p.run = nil
	}
	return err
}

func (p *pipeline) add(name string, n Node) error {
	if p.nodes == nil {
		p.nodes = make(map[string]*container)
	}
	_, ok := p.nodes[name]
	if ok {
		return errors.New("Node exists: " + name)
	}
	p.nodes[name] = &container{name: name, node: n}
	return nil
}

func (p *pipeline) addInput(srcpin string, dstnode *container, dstpin string) error {
	if dstnode == nil || dstnode.name != "args" {
		return unsupportedInputErr
	}
	p.ins = append(p.ins, connection{srcpin, dstnode, dstpin})
	return nil
}

// validate() verifies that the graph is valid.
func (p *pipeline) validate() error {
	descriptions := make(map[string]NodeDescr)
	for _, n := range p.nodes {
		descriptions[n.name] = n.node.Describe()
	}

	for _, n := range p.nodes {
		for _, con := range n.inputs {
			// Input that points to my args list has different rules
			if con.dstNode != nil && con.dstNode.name == "args" {
				continue
			}

			if con.dstNode == nil || con.dstNode.node == nil {
				return errors.New("Node " + n.name + " has no destination for input pin " + con.srcPin)
			}
			// Input that points to the pipeline will be empty, so don't try to validate.
			if con.dstNode.name != "" {
				d, ok := descriptions[con.dstNode.name]
				if !ok {
					return errors.New("Node " + n.name + " has no description for input pin " + con.srcPin + " to " + con.dstNode.name)
				}
				pin := d.FindOutput(con.dstPin)
				if pin == nil {
					return errors.New("Node " + n.name + " has invalid input " + con.srcPin + " to " + con.dstNode.name + ":" + con.dstPin)
				}
			}
		}
		for _, con := range n.outputs {
			if con.dstNode == nil || con.dstNode.node == nil {
				return errors.New("Node " + n.name + " has no destination for output pin " + con.srcPin)
			}
			d, ok := descriptions[con.dstNode.name]
			if !ok {
				return errors.New("Node " + n.name + " has no description for output pin " + con.srcPin + " to " + con.dstNode.name)
			}
			pin := d.FindInput(con.dstPin)
			if pin == nil {
				return errors.New("Node " + n.name + " has invalid output " + con.srcPin + " to " + con.dstNode.name + ":" + con.dstPin)
			}
		}
	}
	return nil
}

func (p *pipeline) gatherInputs(args RunArgs, src Pins) (runnerInput, error) {
	ri := runnerInput{}

	if src != nil && src.Count() > 0 {
		for _, descr := range p.inputDescr {
			docs := src.Remove(descr.Name)
			if docs != nil {
				for _, conn := range descr.connections {
					ri.add(conn.DstNode, conn.DstPin, docs)
				}
			}
		}
	}
	// Apply arguments. We have to find all nodes that have argument inputs.
	for _, dstn := range p.nodes {
		for _, conn := range dstn.inputs {
			if conn.dstNode.name == "args" {
				doc := p.args.valueDoc(args, conn.dstPin)
				if doc != nil {
					docs := []*Doc{doc}
					ri.add(dstn.name, conn.srcPin, docs)
				}
			}
		}
	}
	return ri, nil
}

// gatherSources() returns all source nodes. A source node is one that either
// has no input, or has input data for all its inputs.
func (p *pipeline) gatherSources(input runnerInput) ([]*container, error) {
	if p.nodes == nil {
		return nil, missingSourcesErr
	}
	var sources []*container
	for _, c := range p.nodes {
		if len(c.inputs) < 1 {
			sources = append(sources, c)
		} else if len(input.nodeInputs) > 0 {
			ready := true
			for _, conn := range c.inputs {
				if !input.hasPin(c.name, conn.srcPin) {
					ready = false
					break
				}
			}
			if ready {
				sources = append(sources, c)
			}
		}
	}
	if len(sources) < 1 {
		return nil, missingSourcesErr
	}
	return sources, nil
}

// --------------------------------
// PIPELINE-ARGS

type arg_format string

const (
	empty_format  arg_format = ""
	string_format arg_format = "string"
)

// pipeline_args provides a list of all pipeline args.
type pipeline_args struct {
	env  string
	args map[string]pipeline_arg
}

func (p *pipeline_args) make(frmt arg_format, all map[string]interface{}) error {
	if all == nil {
		return nil
	}
	for k, _v := range all {
		_, exists := p.arg(k)
		if exists {
			return errors.New("Duplicate arg: " + k)
		}
		switch v := _v.(type) {
		case string:
			p.setArg(k, string_format, v)
		}
	}
	return nil
}

// valueDoc() answers a new doc on the given arg, resolving input sources.
func (p *pipeline_args) valueDoc(args RunArgs, name string) *Doc {
	a, ok := p.arg(name)
	if !ok {
		return nil
	}
	// Make env name
	env_name := ""
	if p.env != "" {
		env_name = p.env + strings.ToUpper(name)
	}
	cla_name := name
	return a.valueDoc(args, env_name, cla_name)
}

func (p *pipeline_args) arg(name string) (pipeline_arg, bool) {
	if p.args == nil {
		return pipeline_arg{}, false
	}
	ans, ok := p.args[name]
	return ans, ok
}

func (p *pipeline_args) setArg(name string, format arg_format, value string) {
	if p.args == nil {
		p.args = make(map[string]pipeline_arg)
	}
	p.args[name] = pipeline_arg{format, value}
}

// --------------------------------
// PIPELINE-ARG

// pipeline_arg is a single argument into the pipeline.
type pipeline_arg struct {
	format arg_format
	value  string
}

// valueDoc() answers a new doc on the given arg, resolving input sources.
func (p pipeline_arg) valueDoc(args RunArgs, env_name, cla_name string) *Doc {
	// Order of precedence: default value, environment variable, command line arg.
	value := p.value
	if env_name != "" {
		if v, ok := os.LookupEnv(env_name); ok {
			value = v
		}
	}
	cla := args.ClaValue(cla_name)
	if cla != "" {
		value = cla
	}

	// Note: Currently ignoring the possibility of other types, since we only have string.
	doc := &Doc{MimeType: texttype}
	doc.NewPage("").AddItem(value)
	return doc
}

// --------------------------------
// CONTAINER

// container decorates a node with the input and output connections.
type container struct {
	name    string
	node    Node
	inputs  []connection
	outputs []connection
}

var (
	args_container = &container{name: "args"}
)

func (c *container) connect(srcpin string, dstnode *container, dstpin string) error {
	if dstnode == nil || dstnode.node == nil {
		return BadRequestErr
	}
	c.outputs = append(c.outputs, connection{srcpin, dstnode, dstpin})
	dstnode.inputs = append(dstnode.inputs, connection{dstpin, c, srcpin})
	return nil
}

// connectInput creates a one-way connection, for when we don't have the other node.
func (c *container) connectInput(srcpin string, dstnode *container, dstpin string) error {
	if dstnode == nil || dstnode.name == "" {
		return BadRequestErr
	}
	c.inputs = append(c.inputs, connection{srcpin, dstnode, dstpin})
	return nil
}

// close() closes the node, if possible.
func (c *container) close() error {
	if closer, ok := c.node.(io.Closer); ok && closer != nil {
		return closer.Close()
	}
	return nil
}

// --------------------------------
// CONNECTION

// connection is a single connection from the output of one pin
// to the input of another. It is used to describe both input and output
// connections from each node, but src and dst reverses on input. That
// is, as an output connection, the srcPin is the pin sending data.
// As an input connection, the srcPin is the pin receiving data.
type connection struct {
	srcPin  string
	dstNode *container
	dstPin  string
}

// --------------------------------
// PIPELINE-PIN-DESCR

type pipelinePinDescr struct {
	PinDescr
	connections []connectionDescr
}

type connectionDescr struct {
	DstNode string
	DstPin  string
}

func newConnectionDescr(comp string) connectionDescr {
	v := strings.Split(comp, ":")
	if len(v) != 2 {
		return connectionDescr{}
	}
	return connectionDescr{v[0], v[1]}
}

func pipelineFakeFmt() {
	fmt.Println()
}
