package phly

import (
	"fmt"
	"github.com/micro-go/parse"
)

const (
	batch_txtoutput = "out"
)

var (
	factory_descr = NodeDescr{Id: "phly/batch", Name: "Batch", Purpose: "Perform multiple actions in parallel."}
)

// batch performs batch input in parallel.
type batch struct {
	Count string `json:"count,omitempty"`
	// Pipeline cache
	args  InstantiateArgs `json:"-"`
	cfg   interface{}     `json:"-"`
	descr NodeDescr       `json:"-"`
}

func (n *batch) Describe() NodeDescr {
	// When I'm an instantiated node, I create an instance-accurate descr.
	if n.descr.Name != "" {
		return n.descr
	}
	// When I'm the factory, I create a default descr.
	return factory_descr
}

func (n *batch) Instantiate(args InstantiateArgs, cfg interface{}) (Node, error) {
	_p := &pipeline{}
	p, err := _p.Instantiate(args, cfg)
	if err != nil {
		return nil, err
	}
	descr := p.Describe()
	descr.Id = factory_descr.Id
	descr.Name = factory_descr.Name
	descr.Purpose = factory_descr.Purpose

	return &batch{args: args, cfg: cfg, descr: descr}, nil
}

func (n *batch) Run(args RunArgs, input, output Pins) error {
	count, err := n.getCount(args)
	fmt.Println("RUN BATCH", count, err)
	return nil
}

func (n *batch) getCount(args RunArgs) (int, error) {
	s := args.Env.ReplaceVars(n.Count)
	return parse.SolveInt(s)
}
