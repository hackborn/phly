package phly

import (
	"os"
)

const (
	batch_txtoutput = "out"
)

// batch performs batch input in parallel.
type batch struct {
	Value string `json:"value,omitempty"`
	Cla   string `json:"cla,omitempty"`
	Env   string `json:"env,omitempty"`
}

func (n *batch) Describe() NodeDescr {
	descr := NodeDescr{Id: "phly/batch", Name: "Batch", Purpose: "Perform multiple actions in parallel."}
	//	descr.Cfgs = append(descr.Cfgs, CfgDescr{Name: "value", Purpose: "A value directly entered into the cfg file. Use this if no cla or env are present."})
	//	descr.Cfgs = append(descr.Cfgs, CfgDescr{Name: "env", Purpose: "A value from the environment variables. Use this if no cla is available."})
	//	descr.Cfgs = append(descr.Cfgs, CfgDescr{Name: "cla", Purpose: "A value from the command line arguments."})
	//	descr.OutputPins = append(descr.OutputPins, PinDescr{Name: text_txtoutput, Purpose: "The text output."})
	return descr
}

func (n *batch) Instantiate(args InstantiateArgs, cfg interface{}) (Node, error) {
	return &batch{}, nil
}

func (n *batch) Run(args RunArgs, input, output Pins) error {
	// Order of precedence: default, environment variable, command line arg.
	value := n.Value
	if n.Env != "" {
		if v, ok := os.LookupEnv(n.Env); ok {
			value = v
		}
	}
	cla := args.ClaValue(n.Cla)
	if cla != "" {
		value = cla
	}

	doc := &Doc{MimeType: texttype}
	doc.NewPage("").AddItem(value)
	output.Add(text_txtoutput, doc)
	return nil
}
