package phly

import (
	"testing"
)

func RunNodeTest(t *testing.T, n Node, input Pins) (Pins, error) {
	ts := &TestSender{}
	env = &environment{}
	args := RunArgs{Env: env, Fields: make(map[string]interface{}), cla: make(map[string]string), stop: make(chan struct{})}
	_, err := n.Run(args, input, ts)
	err = MergeErrors(err, ts.err)
	return ts.pins, err
}

// --------------------------------
// SENDER

type TestSender struct {
	pins Pins
	err  error
}

func (t *TestSender) SendPins(n Node, p Pins) {
	t.pins = p
}

func (t *TestSender) SendFinished(n Node, err error) {
	t.err = err
}
