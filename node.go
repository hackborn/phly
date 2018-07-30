package phly

import ()

// --------------------------------
// NODE

// Node performs abstract document processing.
type Node interface {
	Describe() NodeDescr
	Run(args RunArgs, input, output Pins) error
}

// --------------------------------
// NODE-FACTORY

// NodeFactory instantiates a node. All nodes should have their
// factories installed in the bootstrap.
type NodeFactory interface {
	Describe() NodeDescr
	Instantiate(tree interface{}) (Node, error)
}

// --------------------------------
// RUN-ARGS

// RunArgs provides arguments to the node during the run.
type RunArgs struct {
	Fields map[string]interface{}
}
