package phly

import ()

// Node performs abstract document processing.
type Node interface {
	Run(args RunArgs, input, output Pins) error
}

// NodeFactory instantiates a node. All nodes should have their
// factories installed in the bootstrap.
type NodeFactory interface {
	Instantiate(tree interface{}) (Node, error)
}

// RunArgs provides arguments to the node.
type RunArgs struct {
	Fields map[string]interface{}
}
