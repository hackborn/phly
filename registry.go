package phly

import (
	"encoding/json"
	"errors"
)

func Register(name string, fac NodeFactory) error {
	return reg.register(name, fac)
}

var (
	reg = newRegistry()
)

type registry struct {
	factories map[string]NodeFactory
}

func newRegistry() registry {
	factories := make(map[string]NodeFactory)
	return registry{factories}
}

func (r *registry) register(name string, fac NodeFactory) error {
	r.factories[name] = fac
	return nil
}

func (r *registry) instantiate(name string, cfg interface{}) (Node, error) {
	fac, ok := r.factories[name]
	if !ok {
		return nil, errors.New("Missing node: " + name)
	}
	n, err := fac.Instantiate(cfg)
	if err != nil {
		return nil, err
	}
	// Default installation of settings. Nodes can use the standard
	// json tagging to participate, or not to turn it off.
	if cfg != nil {
		b, err := json.Marshal(cfg)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, n)
	}
	return n, err
}
