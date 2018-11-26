package phly

/*
import (
	"fmt"
)

const (
	console_in  = "in"
	console_out = "out"
)

// console logs all input to the console and passes it along.
type console struct {
}

func (n *console) Describe() NodeDescr {
	descr := NodeDescr{Id: "phly/console", Name: "Console", Purpose: "Log all input and pass it on."}
	descr.InputPins = append(descr.InputPins, PinDescr{Name: console_in, Purpose: "Input."})
	descr.OutputPins = append(descr.OutputPins, PinDescr{Name: console_out, Purpose: "Output."})
	return descr
}

func (n *console) Instantiate(args InstantiateArgs, cfg interface{}) (Node, error) {
	return &console{}, nil
}

func (n *console) Run(args RunArgs, input Pins, sender PinSender) (Flow, error) {
	input.Walk(func(name string, docs []*Doc) {
		fmt.Println(name)
		if len(docs) > 0 {
			for _, doc := range docs {
				fmt.Println("doc pages:", len(doc.Pages))
				if len(doc.Pages) > 0 {
					for _, page := range doc.Pages {
						fmt.Println("\tpage", page.Name)
						if len(page.Items) > 0 {
							for _, item := range page.Items {
								fmt.Println("\t\t", item)
							}
						}
					}
				}
			}
		}
	})

	sender.SendPins(n, input)
	return Running, nil
}
*/
