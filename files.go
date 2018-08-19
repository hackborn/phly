package phly

import (
	"os"
	"path/filepath"
	"strings"
)

const (
	files_input  = "in"
	files_output = "out"
)

// files creates a file list from file and folder names.
type files struct {
	Sep     string `json:"sep,omitempty"`
	Expand  bool   `json:"expand,omitempty"`
	Recurse bool   `json:"recurse,omitempty"`
}

func (n *files) Describe() NodeDescr {
	descr := NodeDescr{Id: "phly/files", Name: "Files", Purpose: "Create file lists from file names and folders. Produce a single doc with a single page."}
	descr.Cfgs = append(descr.Cfgs, CfgDescr{Name: "sep", Purpose: "A separator character. Used to split incoming strings into multiple file paths."})
	descr.Cfgs = append(descr.Cfgs, CfgDescr{Name: "expand", Purpose: "(true or false). When true, folders are expanded to the file contents."})
	descr.InputPins = append(descr.InputPins, PinDescr{Name: files_input, Purpose: "The folder or file list."})
	descr.OutputPins = append(descr.OutputPins, PinDescr{Name: files_output, Purpose: "The file list."})
	return descr
}

func (n *files) Instantiate(args InstantiateArgs, cfg interface{}) (Node, error) {
	return &files{}, nil
}

func (n *files) Run(args RunArgs, input Pins, sender PinSender) (Flow, error) {
	doc := &Doc{MimeType: texttype}
	page := doc.NewPage("")
	srcs := input.Get(files_input)
	var err error
	if srcs != nil {
		for _, sdoc := range srcs {
			for _, spage := range sdoc.Pages {
				for _, sitem := range spage.Items {
					err = MergeErrors(err, n.addItem(sitem, page))
				}
			}
		}
	}
	if err != nil {
		return nil, err
	}
	pins := NewPins()
	pins.Add(text_txtoutput, doc)
	sender.SendPins(n, pins)
	return Finished, nil
}

func (n *files) Stop() error {
	return nil
}

func (n *files) addItem(item interface{}, page *Page) error {
	s, ok := item.(string)
	if !ok {
		return BadRequestErr
	}
	list := []string{s}
	if n.Sep != "" {
		list = strings.Split(s, n.Sep)
	}
	var err error
	for _, item := range list {
		if n.Expand {
			err = MergeErrors(err, n.expand(item, page))
		} else {
			page.AddItem(item)
		}
	}
	return err
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
