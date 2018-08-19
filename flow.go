package phly

import ()

type Flow interface {
}

var (
	// The operation is done.
	Finished = &flow{1}

	// Waiting indicates that the node is waiting to receive all input
	// before processing. The framework will accunulate the input, running
	// the node with each new input that arrives. Once all input has arrived
	// and the node stops reporting Waiting it will be removed from the
	// running stack. Any node that must receive all input before processing
	// should use this flow.
	Waiting = &flow{2}

	// Running indicates that the node is running. It should consume input
	// as it arrives. It will not be removed from the running stack until
	// it sends a finished message.
	Running = &flow{3}
)

type flow struct {
	v int
}
