package phly

import (
	"errors"
)

type Pipeline interface {
	Run(args RunArgs) (PipelineResult, error)
}

func NewPipeline() Pipeline {
	return &pipeline{}
}

// --------------------------------
// PIPELINE

type pipeline struct {
	nodes map[string]*container
}

func (p *pipeline) Run(args RunArgs) (PipelineResult, error) {
	sources, err := p.sources()
	if err != nil {
		return PipelineResult{}, err
	}
	r := &runner{}
	_, err = r.run(args, p, sources)
	return PipelineResult{}, err
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
			d, ok := descriptions[con.dstNode.name]
			if !ok {
				return errors.New("Node " + n.name + " has no description for input pin " + con.srcPin + " to " + con.dstNode.name)
			}
			pin := d.FindOutput(con.dstPin)
			if pin == nil {
				return errors.New("Node " + n.name + " has invalid input " + con.srcPin + " to " + con.dstNode.name + ":" + con.dstPin)
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

func (p *pipeline) sources() ([]*container, error) {
	if p.nodes == nil {
		return nil, missingSourcesErr
	}
	var sources []*container
	for _, c := range p.nodes {
		if len(c.inputs) < 1 {
			sources = append(sources, c)
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
// to the input of another. It is to describe both input and output
// connections from each node, but src and dst reverses on input.
type connection struct {
	srcPin  string
	dstNode *container
	dstPin  string
}

// --------------------------------
// PIPELINE-RESULT

type PipelineResult struct {
}
