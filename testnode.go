package phly

import (
	"strings"
)

const (
	testnode_in  = "in"
	testnode_out = "out"
)

// testnode is used solely in tests.
type testnode struct {
	RunMode string   `json:"runmode,omitempty"`
	Finish  []string `json:"finish,omitempty"`
}

func (n *testnode) Describe() NodeDescr {
	descr := NodeDescr{Id: "phly/test", Name: "Test", Purpose: "A node for running tests."}
	descr.InputPins = append(descr.InputPins, PinDescr{Name: testnode_in, Purpose: "Input."})
	descr.OutputPins = append(descr.OutputPins, PinDescr{Name: testnode_out, Purpose: "Output."})
	return descr
}

func (n *testnode) Instantiate(args InstantiateArgs, cfg interface{}) (Node, error) {
	return &testnode{}, nil
}

func (n *testnode) Run(args RunArgs, input Pins, sender PinSender) (Flow, error) {
	runmode := n.runMode()
	if runmode == Finished {
		n.sendPinsFrom(n.Finish, sender)
	}
	return runmode, nil
}

func (n *testnode) sendPinsFrom(items []string, sender PinSender) {
	pins := n.pinsFrom(items)
	if pins != nil && pins.Count() > 0 {
		sender.SendPins(n, pins)
	}
}

func (n *testnode) pinsFrom(items []string) Pins {
	pins := NewPins()
	if len(items) > 0 {
		doc := &Doc{}
		page := doc.NewPage("")
		for _, i := range items {
			page.AddItem(i)
		}
		pins.Add(testnode_out, doc)
	}
	return pins
}

func (n *testnode) runMode() Flow {
	if n.RunMode != "" {
		rm := strings.ToLower(n.RunMode)
		if strings.HasPrefix(rm, "f") {
			return Finished
		} else if strings.HasPrefix(rm, "r") {
			return Running
		}
	}
	return Finished
}
