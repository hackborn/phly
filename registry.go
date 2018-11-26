package phly

import (
	"encoding/json"
)

func Register(fac NodeFactory) error {
	return reg.register(fac)
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

func (r *registry) register(fac NodeFactory) error {
	id := fac.Describe().Id
	if id == "" {
		return NewBadRequestError("NodeFactory missing ID")
	}
	r.factories[id] = fac
	return nil
}

func (r *registry) instantiate(name string, cfg interface{}) (Node, error) {
	fac, ok := r.factories[name]
	if !ok {
		return nil, NewMissingError("Node " + name)
	}
	args := InstantiateArgs{Env: env}
	n, err := fac.Instantiate(args, cfg)
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
