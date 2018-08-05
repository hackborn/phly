package phly

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	files_output = "out"
)

// files creates a file list from file and folder names.
type files struct {
	Value  string `json:"value,omitempty"`
	Cla    string `json:"cla,omitempty"`
	Env    string `json:"env,omitempty"`
	Expand bool   `json:"expand,omitempty"`
}

func (n *files) Describe() NodeDescr {
	descr := NodeDescr{Id: "phly/files", Name: "Files", Purpose: "Create file lists from file names and folders. Produce a single doc with a single page."}
	descr.Cfgs = append(descr.Cfgs, CfgDescr{Name: "value", Purpose: "A value directly entered into the cfg file. Use this if no cla or env are present."})
	descr.Cfgs = append(descr.Cfgs, CfgDescr{Name: "env", Purpose: "A value from the environment variables. Use this if no cla is available."})
	descr.Cfgs = append(descr.Cfgs, CfgDescr{Name: "cla", Purpose: "A value from the command line arguments."})
	descr.Cfgs = append(descr.Cfgs, CfgDescr{Name: "expand", Purpose: "(true or false). When true, folders are expanded to the file contents."})
	descr.OutputPins = append(descr.OutputPins, PinDescr{Name: files_output, Purpose: "The file list."})
	return descr
}

func (n *files) Instantiate(args InstantiateArgs, cfg interface{}) (Node, error) {
	return &files{}, nil
}

func (n *files) Run(args RunArgs, input, output Pins) error {
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
	page := doc.NewPage("")
	list := strings.Split(value, ";")
	for _, item := range list {
		if n.Expand {
			err := n.expand(item, page)
			if err != nil {
				return err
			}
		} else {
			page.AddItem(item)
		}
	}
	output.Add(text_txtoutput, doc)
	return nil
}

func (n *files) expand(path string, page *Page) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			page.AddItem(path)
		}
		return nil
	})
	return err
}
