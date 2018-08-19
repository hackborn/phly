package phly

import (
	"fmt"
	"github.com/micro-go/parse"
	"strings"
	"sync"
)

type batch_mode int

const (
	batch_txtoutput            = "out"
	mode_doc        batch_mode = iota
	mode_page
	mode_item
)

var (
	factory_descr = NodeDescr{Id: "phly/batch", Name: "Batch", Purpose: "Perform multiple actions in parallel."}
)

// batch performs batch input in parallel.
type batch struct {
	Mode  string `json:"mode,omitempty"`
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

func (n *batch) Run(args RunArgs, input Pins, sender PinSender) (Flow, error) {
	count, err := n.getCount(args)
	batch_input := n.flattenInput(input)
	if err != nil || batch_input.len() < 1 {
		return nil, err
	}
	// Don't create more threads than we need
	if batch_input.len() < count {
		count = batch_input.len()
	}

	batch_output := shuttle{}
	wg := sync.WaitGroup{}
	err = nil
	for i := 0; i < count; i++ {
		_p := &pipeline{}
		node, err2 := _p.Instantiate(n.args, n.cfg)
		err = MergeErrors(err, err2)
		// Slightly odd, but we don't have a stop in place, so start as
		// many threads even if one errors in hopes we get through the list faster.
		if err2 == nil {
			wg.Add(1)
			go n.run(&wg, node, args.copy(), batch_input, &batch_output, sender)
		}
	}
	wg.Wait()
	return Running, err
}

func (n *batch) Stop() error {
	return nil
}

func (n *batch) run(wg *sync.WaitGroup, node Node, args RunArgs, input *shuttle, output *shuttle, sender PinSender) {
	defer wg.Done()
	for _i, err := input.pop(); err == nil; _i, err = input.pop() {
		if pd, ok := _i.(*pindoc); ok && pd.doc != nil {
			pinin := &pins{}
			pinin.Add(pd.pin, pd.doc)
			_, err := node.Run(args, pinin, sender)
			if err != nil {
				fmt.Println("batch err", err)
			}
		}
	}
}

func (n *batch) getCount(args RunArgs) (int, error) {
	s := args.Env.ReplaceVars(n.Count)
	return parse.SolveInt(s)
}

func (n *batch) flattenInput(input Pins) *shuttle {
	stack := &shuttle{}
	if input == nil {
		return stack
	}
	mode := n.getMode()
	input.Walk(func(name string, docs []*Doc) {
		for _, d := range docs {
			if mode == mode_doc {
				stack.push(&pindoc{name, d})
			} else {
				for _, p := range d.Pages {
					if mode == mode_page {
						dst := &Doc{}
						dst.AddPage(p.Copy())
						stack.push(&pindoc{name, dst})
					} else {
						for _, i := range p.Items {
							dst := &Doc{}
							dst.NewPage("").AddItem(i)
							stack.push(&pindoc{name, dst})
						}
					}
				}
			}
		}
	})
	return stack
}

func (n *batch) getMode() batch_mode {
	if strings.ToLower(n.Mode) == "items" {
		return mode_item
	} else if strings.ToLower(n.Mode) == "pages" {
		return mode_page
	}
	return mode_doc
}

// --------------------------------
// PINDOC

// pindoc is a pin and doc; used to get data to the batch runner.
type pindoc struct {
	pin string
	doc *Doc
}
