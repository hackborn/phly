package phly

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/micro-go/parse"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func LoadPipeline(name string) (Pipeline, error) {
	filename := env.FindFile(name)
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadPipeline(f)
}

func ReadPipeline(r io.Reader) (Pipeline, error) {
	p := &pipeline{}
	err := readPipeline(r, p)
	fmt.Println("READ PIPELNE STOPPED", p.stopped)
	return p, err
}

func readPipeline(r io.Reader, p *pipeline) error {
	p.workingdir = workingDirFrom(r)

	d := json.NewDecoder(r)
	cfg := &pipelinecfg{}
	err := d.Decode(cfg)
	if err != nil {
		return err
	}
	cfg.applyEnvVarsToPins()
	//	fmt.Println("LOADED", cfg)
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

func (p *pipelinecfg) applyEnvVarsToPins() {
	// Replace any pin names with environment variables. Note this is only
	// the names, and used for doing things like allowing different values
	// for different platforms.
	p.applyEnvVarsToNodes()
	applyEnvVarsToSingle(p.Args.Strings)
	applyEnvVarsToMultiple(p.Ins)
	applyEnvVarsToMultiple(p.Outs)
}

func (p *pipelinecfg) applyEnvVarsToNodes() {
	for _, n := range p.Nodes {
		applyEnvVarsToSingle(treeMapStrings("ins", n))
		applyEnvVarsToSingle(treeMapStrings("outs", n))
	}
}

func applyEnvVarsToSingle(m map[string]interface{}) {
	if m == nil {
		return
	}
	for k, v := range m {
		newk := env.ReplaceVars(k, nil)
		newv, changed := applyEnvVarsToInterface(v)
		if changed || newk != k {
			delete(m, k)
			m[newk] = newv
		}
	}
}

func applyEnvVarsToMultiple(m map[string][]string) {
	if m == nil {
		return
	}
	for k, v := range m {
		newk := env.ReplaceVars(k, nil)
		if newk != k {
			delete(m, k)
			m[newk] = v
		}
	}
}

// applyEnvVarsToInterface() applies the environment variables to an unknown type,
// answering the new value and true if it changed.
func applyEnvVarsToInterface(_v interface{}) (interface{}, bool) {
	switch v := _v.(type) {
	case string:
		newv := env.ReplaceVars(v, nil)
		if v != newv {
			return newv, true
		}
	}
	return _v, false
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

func treeMapStrings(path string, tree interface{}) map[string]interface{} {
	if tree == nil {
		return nil
	}
	_v, err := parse.TreeValue(path, tree)
	if err != nil {
		return nil
	}
	if v, ok := _v.(map[string]interface{}); ok {
		return v
	}
	return nil
}

// --------------------------------
// PIPELINE-ARGS-IO

type pipeline_args_io struct {
	Env     string                 `json:"env,omitempty"`
	Strings map[string]interface{} `json:"strings,omitempty"`
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

func workingDirFrom(r io.Reader) string {
	if n, ok := r.(namer); ok && n != nil {
		return filepath.Dir(n.Name())
	}
	return ""
}
