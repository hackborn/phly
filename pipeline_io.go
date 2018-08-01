package phly

import (
	"encoding/json"
	"errors"
	"github.com/micro-go/parse"
	"io"
	"os"
	"strings"
)

func LoadPipeline(filename string) (Pipeline, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return ReadPipeline(f)
}

func ReadPipeline(r io.Reader) (Pipeline, error) {
	d := json.NewDecoder(r)
	cfg := pipelinecfg{}
	err := d.Decode(&cfg)
	if err != nil {
		return nil, err
	}
	if len(cfg.Nodes) < 1 {
		return nil, BadRequestErr
	}
	p := &pipeline{}
	pins := make(map[string][]pincfg)
	// Create the nodes and cache their pins
	for k, v := range cfg.Nodes {
		n, err := readNode(k, v)
		if err != nil {
			return nil, err
		}
		err = p.add(k, n)
		if err != nil {
			return nil, err
		}
		pc, err := readPinCfgs(v)
		if err != nil {
			return nil, err
		}
		if pc != nil {
			pins[k] = pc
		}
	}
	// Apply the pins
	for k, pinlist := range pins {
		srcn, ok := p.nodes[k]
		if !ok || srcn == nil {
			return nil, errors.New("Pin on missing node " + k)
		}
		for _, pin := range pinlist {
			dstn, ok := p.nodes[pin.dstNode]
			if !ok || dstn == nil {
				return nil, errors.New("Pin on missing node " + pin.dstNode)
			}
			err = srcn.connect(pin.srcPin, dstn, pin.dstPin)
			if err != nil {
				return nil, err
			}
		}
	}
	// Validate
	err = p.validate()
	return p, err
}

// --------------------------------
// PIPELINE-CFG

type pipelinecfg struct {
	Nodes map[string]interface{} `json:"nodes,omitempty"`
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
	cfg, _ := parse.FindTreeValue("cfg", v)
	n, err := reg.instantiate(name, cfg)
	return n, err
}

func readPinCfgs(v interface{}) ([]pincfg, error) {
	_pc, ok := parse.FindTreeValue("pins", v)
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
