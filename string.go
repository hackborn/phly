package phly

import (
	"mime"
	"os"
)

const (
	string_txtoutput = "0"
)

var (
	stringtype = mime.TypeByExtension(".txt")
)

// stringnode gathers a string to send out the pin.
type stringnode struct {
	Value string `json:"value,omitempty"`
	Cla   string `json:"cla,omitempty"`
	Env   string `json:"env,omitempty"`
}

func (n *stringnode) Describe() NodeDescr {
	descr := NodeDescr{Id: "phly/string", Name: "String"}
	descr.OutputPins = append(descr.OutputPins, PinDescr{Name: string_txtoutput, Purpose: "The single string output."})
	return descr
}

func (n *stringnode) Run(args RunArgs, input, output Pins) error {
	// Order of precedence: default, environment variable, command line arg.
	value := n.Value
	if n.Env != "" {
		value = os.Getenv(n.Env)
	}

	doc := &Doc{MimeType: stringtype}
	doc.NewPage("").AddItem(value)
	output.Add(string_txtoutput, doc)
	return nil
}

func (n *stringnode) Instantiate(cfg interface{}) (Node, error) {
	return &stringnode{}, nil
}
