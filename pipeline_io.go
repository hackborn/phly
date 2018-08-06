package phly

import (
	"encoding/json"
	"errors"
	"github.com/micro-go/parse"
	"io"
	"os"
	"strings"
)

func LoadPipeline(name string) (Node, error) {
	filename := env.FindFile(name)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadPipeline(f)
}

func ReadPipeline(r io.Reader) (Node, error) {
	p := &pipeline{}
	err := readPipeline(r, p)
	return p, err
}

func readPipeline(r io.Reader, p *pipeline) error {
	d := json.NewDecoder(r)
	cfg := pipelinecfg{}
	err := d.Decode(&cfg)
	if err != nil {
		return err
	}
	// fmt.Println("LOADED", cfg)
	if len(cfg.Nodes) < 1 {
		return BadRequestErr
	}
	p.inputDescr = makePipelinePinDescrs(cfg.Ins)
	p.outputDescr = makePipelinePinDescrs(cfg.Outs)
	node_ins := make(map[string][]pincfg)
	node_outs := make(map[string][]pincfg)

	// Read the args
	args, err := cfg.Args.asArgs()
	if err != nil {
		return err
	}
	p.args = args

	// Create the nodes and cache their pins
	for k, v := range cfg.Nodes {
		n, err := readNode(k, v)
		if err != nil {
			return err
		}
		err = MergeErrors(err, p.add(k, n))
		err = MergeErrors(err, readPinCfgsTo("outs", v, k, node_outs))
		err = MergeErrors(err, readPinCfgsTo("ins", v, k, node_ins))
		if err != nil {
			return err
		}
	}

	// Apply the out pins
	for k, pinlist := range node_outs {
		srcn, ok := p.nodes[k]
		if !ok || srcn == nil {
			return errors.New("Pin on missing node " + k)
		}
		for _, pin := range pinlist {
			dstn, ok := p.nodes[pin.dstNode]
			if !ok || dstn == nil {
				return errors.New("Pin on missing node " + pin.dstNode)
			}
			err = srcn.connect(pin.srcPin, dstn, pin.dstPin)
			if err != nil {
				return err
			}
		}
	}

	// Apply the in pins
	// NOTE: Currently just used for connecting pipelines to the args list.
	for k, pinlist := range node_ins {
		srcn, ok := p.nodes[k]
		if !ok || srcn == nil {
			return errors.New("Pin on missing node " + k)
		}
		for _, pin := range pinlist {
			if pin.dstNode == "args" {
				err = MergeErrors(err, p.addInput(pin.srcPin, args_container, pin.dstPin))
				err = MergeErrors(err, srcn.connectInput(pin.srcPin, args_container, pin.dstPin))
			} else {
				return errors.New("Invalid in pin on " + pin.dstNode + " - ins are limited to args")
			}
		}
	}
	if err != nil {
		return err
	}

	// Hookup my inputs. They point to an empty container, since there's no
	// container for the pipeline. Inputs aren't traversed so this is fine.
	empty_container := &container{node: p}
	for _, descr := range p.inputDescr {
		for _, conn := range descr.connections {
			dstn, ok := p.nodes[conn.DstNode]
			if !ok || dstn == nil {
				return errors.New("Pipeline input pin on missing node " + conn.DstNode)
			}
			dstn.inputs = append(dstn.inputs, connection{conn.DstPin, empty_container, descr.Name})
		}
	}

	// Validate
	return p.validate()
}

// --------------------------------
// PIPELINE-CFG

type pipelinecfg struct {
	Args  pipeline_args_io       `json:"args,omitempty"`
	Ins   map[string][]string    `json:"ins,omitempty"`
	Outs  map[string][]string    `json:"outs,omitempty"`
	Nodes map[string]interface{} `json:"nodes,omitempty"`
}

func makePipelinePinDescrs(src map[string][]string) []pipelinePinDescr {
	var dst []pipelinePinDescr
	for k, v := range src {
		descr := pipelinePinDescr{}
		descr.Name = k
		for _, vv := range v {
			descr.connections = append(descr.connections, newConnectionDescr(vv))
		}
		dst = append(dst, descr)
	}
	return dst
}

// --------------------------------
// PIPELINE-ARGS-IO

type pipeline_args_io struct {
	Env     string            `json:"env,omitempty"`
	Strings map[string]string `json:"strings,omitempty"`
}

func (p pipeline_args_io) asArgs() (pipeline_args, error) {
	a := &pipeline_args{env: p.Env}
	err := a.make(string_format, p.Strings)
	return *a, err
}

// --------------------------------
// PIN-CFG

type pincfg struct {
	srcPin  string
	dstNode string
	dstPin  string
}

// --------------------------------
// MISC

func readNode(k string, v interface{}) (Node, error) {
	name, _ := parse.FindTreeString("node", v)
	if name == "" {
		return nil, errors.New("Missing node for " + k)
	}
	if !isLegalNodeName(name) {
		return nil, errors.New("Illegal node name: " + name)
	}
	cfg, _ := parse.FindTreeValue("cfg", v)
	n, err := reg.instantiate(name, cfg)
	return n, err
}

func readPinCfgsTo(name string, v interface{}, dstkey string, dst map[string][]pincfg) error {
	pc, err := readPinCfgs(name, v)
	if err != nil {
		return err
	}
	if pc != nil {
		dst[dstkey] = pc
	}
	return nil
}

func readPinCfgs(name string, v interface{}) ([]pincfg, error) {
	_pc, ok := parse.FindTreeValue(name, v)
	if !ok {
		// No pins are valid
		return nil, nil
	}
	pc, ok := _pc.(map[string]interface{})
	if !ok {
		return nil, wrongFormatPinsErr
	}
	var ans []pincfg
	for k, _v := range pc {
		v, ok := _v.(string)
		if !ok {
			return nil, wrongFormatPinsErr
		}
		parts := strings.Split(v, ":")
		if len(parts) != 2 {
			return nil, wrongFormatPinsErr
		}
		ans = append(ans, pincfg{k, parts[0], parts[1]})
	}
	return ans, nil
}

func isLegalNodeName(name string) bool {
	switch strings.ToLower(name) {
	case
		"ins", "outs", "args":
		return false
	}
	return true
}
