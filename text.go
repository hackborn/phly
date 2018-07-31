package phly

import (
	"mime"
	"os"
)

const (
	text_txtoutput = "0"
)

var (
	texttype = mime.TypeByExtension(".txt")
)

// text gathers a string to send out the pin.
type text struct {
	Value string `json:"value,omitempty"`
	Cla   string `json:"cla,omitempty"`
	Env   string `json:"env,omitempty"`
}

func (n *text) Describe() NodeDescr {
	descr := NodeDescr{Id: "phly/text", Name: "Text", Purpose: "Acquire text from the cfg values. If a cla is available use that. If no cla, use the env. If no env, use the value."}
	descr.Cfgs = append(descr.Cfgs, CfgDescr{Name: "value", Purpose: "A value directly entered into the cfg file. Use this if no cla or env are present."})
	descr.Cfgs = append(descr.Cfgs, CfgDescr{Name: "env", Purpose: "A value from the environment variables. Use this if no cla is available."})
	descr.Cfgs = append(descr.Cfgs, CfgDescr{Name: "cla", Purpose: "A value from the command line arguments."})
	descr.OutputPins = append(descr.OutputPins, PinDescr{Name: text_txtoutput, Purpose: "The text output."})
	return descr
}

func (n *text) Run(args RunArgs, input, output Pins) error {
	// Order of precedence: default, environment variable, command line arg.
	value := n.Value
	if n.Env != "" {
		if v, ok := os.LookupEnv(n.Env); ok {
			value = v
		}
	}

	doc := &Doc{MimeType: texttype}
	doc.NewPage("").AddItem(value)
	output.Add(text_txtoutput, doc)
	return nil
}

func (n *text) Instantiate(cfg interface{}) (Node, error) {
	return &text{}, nil
}
