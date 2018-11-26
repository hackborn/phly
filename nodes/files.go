package phly_nodes

import (
	"github.com/hackborn/phly"
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

func (n *files) Describe() phly.NodeDescr {
	descr := phly.NodeDescr{Id: "phly/files", Name: "Files", Purpose: "Create file lists from file names and folders. Produce a single doc with a single page."}
	descr.Cfgs = append(descr.Cfgs, phly.CfgDescr{Name: "sep", Purpose: "A separator character. Used to split incoming strings into multiple file paths."})
	descr.Cfgs = append(descr.Cfgs, phly.CfgDescr{Name: "expand", Purpose: "(true or false). When true, folders are expanded to the file contents."})
	descr.InputPins = append(descr.InputPins, phly.PinDescr{Name: files_input, Purpose: "The folder or file list."})
	descr.OutputPins = append(descr.OutputPins, phly.PinDescr{Name: files_output, Purpose: "The file list."})
	return descr
}

func (n *files) Instantiate(args phly.InstantiateArgs, cfg interface{}) (phly.Node, error) {
	return &files{}, nil
}

func (n *files) Process(args phly.ProcessArgs, stage phly.NodeStage, input phly.Pins, output phly.NodeOutput) error {
	var err error
	doc := &phly.Doc{MimeType: texttype}
	phly.WalkStringItems(input, files_input, func(channel string, src *phly.Doc, index int, item string) {
		err = phly.MergeErrors(err, n.addItem(item, doc))
	})
	if len(doc.Items) > 0 {
		output.SendPins(phly.PinBuilder{}.Add(files_output, doc).Pins())
	}
	// Run once and we're done
	output.SendMsg(phly.MsgFromStop(nil))
	return err
}

func (n *files) StopNode(args phly.StoppedArgs) error {
	return nil
}

func (n *files) addItem(s string, dst *phly.Doc) error {
	list := []string{s}
	if n.Sep != "" {
		list = strings.Split(s, n.Sep)
	}
	var err error
	for _, item := range list {
		if n.Expand {
			err = phly.MergeErrors(err, n.expand(item, dst))
		} else {
			dst.AppendItem(item)
		}
	}
	return err
}

func (n *files) expand(path string, dst *phly.Doc) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			dst.AppendItem(path)
		}
		return nil
	})
	return err
}
