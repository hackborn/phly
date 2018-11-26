package phly

import ()

type NodeStage string

const (
	NodeStarting NodeStage = "starting"
	NodeRunning            = "running"
)

// ----------------------------------------
// NODE

// Node performs abstract document processing.
type Node interface {
	Describe() NodeDescr

	// Process processes input, sending any results to output.
	// stage provides minimal lifecycle information: whether this node is starting or not.
	// input is the collection of inputs to process.
	// output is used to send any output to the calling graph. It is also used to
	// request StopNode() be called if processing should end.
	// error is used to report any errors; not that it will immediately stop all
	// graph processing.
	Process(args ProcessArgs, stage NodeStage, input Pins, output NodeOutput) error

	// Request to stop the node.
	StopNode(args StoppedArgs) error
}

// ----------------------------------------
// NODE-FACTORY

// NodeFactory instantiates a node. All nodes should have their
// factories installed in the bootstrap.
type NodeFactory interface {
	Describe() NodeDescr
	Instantiate(args InstantiateArgs, tree interface{}) (Node, error)
}

// ----------------------------------------
// NODE-OUTPUT

// NodeOutput sends data from the node to its parent runner.
type NodeOutput interface {
	SendPins(Pins)
	SendMsg(Msg)
}
