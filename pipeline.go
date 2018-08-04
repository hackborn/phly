package phly

import (
	"errors"
	"github.com/micro-go/parse"
	"strings"
)

// --------------------------------
// PIPELINE

type pipeline struct {
	file        string                `json:"-"`
	nodes       map[string]*container `json:"-"`
	inputDescr  []pipelinePinDescr    `json:"-"`
	outputDescr []pipelinePinDescr    `json:"-"`
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

func (p *pipeline) Run(args RunArgs, input, output Pins) error {
	// Make my input and sources
	inputs, err := p.gatherInputs(input)
	if err != nil {
		return err
	}
	sources, err := p.gatherSources(inputs)
	if err != nil {
		return err
	}
	r := &runner{}
	outs, err := r.run(args, p, sources, inputs)
	if outs != nil {
		outs.Describe()
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

// validate() verifies that the graph is valid.
func (p *pipeline) validate() error {
	descriptions := make(map[string]NodeDescr)
	for _, n := range p.nodes {
		descriptions[n.name] = n.node.Describe()
	}

	for _, n := range p.nodes {
		for _, con := range n.inputs {
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

func (p *pipeline) gatherInputs(src Pins) (runnerInput, error) {
	ri := runnerInput{}
	if src == nil || src.Count() < 1 {
		return ri, nil
	}
	// Each input gets turned into an entry pointing to one of my nodes.
	for _, descr := range p.inputDescr {
		docs := src.Get(descr.Name)
		if docs != nil {
			for _, conn := range descr.connections {
				ri.add(conn.DstNode, conn.DstPin, docs)
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
// CONTAINER

// container decorates a node with the input and output connections.
type container struct {
	name    string
	node    Node
	inputs  []connection
	outputs []connection
}

func (c *container) connect(srcpin string, dstnode *container, dstpin string) error {
	if dstnode == nil || dstnode.node == nil {
		return BadRequestErr
	}
	c.outputs = append(c.outputs, connection{srcpin, dstnode, dstpin})
	dstnode.inputs = append(dstnode.inputs, connection{dstpin, c, srcpin})
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
