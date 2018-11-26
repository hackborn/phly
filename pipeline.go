package phly

import (
	"errors"
	"fmt"
	"github.com/micro-go/lock"
	"github.com/micro-go/parse"
	"io"
	"os"
	"strings"
	"sync"
)

// --------------------------------
// PIPELINE interface

type Pipeline interface {
	Run(args StartArgs, input Pins) error
	Start(args StartArgs, input Pins) error
	Stop() error
	Wait() error
}

// --------------------------------
// PIPELINE start

// StartArgs provides arguments when starting the pipeline.
type StartArgs struct {
	Cla    map[string]string // Command line arguments
	output NodeOutput        // The receiver for any output from this pipeline
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
	mutex  sync.Mutex       `json:"-"`
	runner *pipeline_runner `json:"-"`
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

func (p *pipeline) Process(args ProcessArgs, stage NodeStage, input Pins, output NodeOutput) error {
	if stage == NodeStarting {
		p.Stop()
		// XXX I guess I need to cache the node output or something -- how do I get data out?
		sargs := StartArgs{Cla: args.cla, output: output}
		fmt.Println("PROCESS CALLING START")
		return p.Start(sargs, input)
	}
	return nil
}

func (p *pipeline) StopNode(args StoppedArgs) error {
	return p.Stop()
}

func (p *pipeline) Run(args StartArgs, input Pins) error {
	err := p.Start(args, input)
	if err != nil {
		return err
	}
	return p.Wait()
}

func (p *pipeline) Start(args StartArgs, input Pins) error {
	p.Stop()
	pargs := ProcessArgs{env: env, workingdir: p.workingdir, cla: args.Cla}

	defer lock.Locker(&p.mutex).Unlock()
	runner, err := startPipelineRunner(p, args, pargs, input)
	if err != nil {
		return err
	}
	fmt.Println("pipeline started")
	p.runner = runner
	return nil
}

func (p *pipeline) Stop() error {
	defer lock.Locker(&p.mutex).Unlock()

	var err error
	if p.runner != nil {
		err = p.runner.close()
		p.runner = nil
	}
	return err
}

func (p *pipeline) Wait() error {
	r := p.getRunner()
	if r == nil {
		return NewIllegalError("Waiting but nothing started")
	}
	fmt.Println("START WAIT")
	r.wait.Wait()
	fmt.Println("STOP WAIT err", r.err.Get())
	return r.err.Get()
}

// getRunner() answers the current runner, so we don't have
// to keep a lock during the wait.
func (p *pipeline) getRunner() *pipeline_runner {
	defer lock.Locker(&p.mutex).Unlock()
	return p.runner
}

// ResolveOutput() takes a source node name and pin and converts it into the
// destination node name and pin.
func (p *pipeline) ResolveOutput(srcnode, srcpin string) (string, string, error) {
	c, ok := p.nodes[srcnode]
	if c == nil || !ok {
		return "", "", NewBadRequestError("Node " + srcnode + " does not exist")
	}
	for _, conn := range c.outputs {
		if conn.srcPin == srcpin {
			return conn.dstNode.name, conn.dstPin, nil
		}
	}
	return "", "", NewBadRequestError("Node " + srcnode + " does not have pin " + srcpin)
}

func (p *pipeline) add(name string, n Node) error {
	if p.nodes == nil {
		p.nodes = make(map[string]*container)
	}
	_, ok := p.nodes[name]
	if ok {
		return NewIllegalError("Duplicate nodes named " + name)
	}
	p.nodes[name] = &container{name: name, node: n}
	return nil
}

func (p *pipeline) addInput(srcpin string, dstnode *container, dstpin string) error {
	if !(dstnode.name == args_container.name || dstnode.name == pipeline_container.name) {
		return NewIllegalError("Invalid pin " + srcpin + " to " + dstnode.name + "->" + dstpin)
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
			// Input that points to my hardcoded inputs have different rules
			if con.dstNode != nil && (con.dstNode.name == args_container.name || con.dstNode.name == pipeline_container.name) {
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
func (p *pipeline_args) valueDoc(args ProcessArgs, name string) *Doc {
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
func (p pipeline_arg) valueDoc(args ProcessArgs, env_name, cla_name string) *Doc {
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
	doc.AppendItem(value)
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
	args_container     = &container{name: "args"}
	pipeline_container = &container{name: ".pipeline"}
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

// --------------------------------
// STOPPED

const (
	emptyStopped       = iota // The pipeline has not been started
	running                   // The pipeline is running
	requestedStopped          // A client has requested the pipeline to stop
	runFinishedStopped        // The pipeline processing finished
)
