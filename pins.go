package phly

import (
	"fmt"
)

type PinsWalkFunc func(name string, docs []*Doc)

// Pins describes a collection of documents attached to pins.
type Pins interface {
	Count() int
	Get(name string) []*Doc
	Add(name string, doc *Doc)
	Remove(name string) []*Doc
	Walk(fn PinsWalkFunc)
	Describe()
}

func NewPins() Pins {
	return &pins{}
}

type pins struct {
	all map[string][]*Doc
}

func (p *pins) Count() int {
	return len(p.all)
}

func (p *pins) Get(name string) []*Doc {
	if p.all == nil {
		return nil
	}
	return p.all[name]
}

func (p *pins) Add(name string, doc *Doc) {
	if p.all == nil {
		p.all = make(map[string][]*Doc)
	}
	slice := p.all[name]
	slice = append(slice, doc)
	p.all[name] = slice
}

func (p *pins) Remove(name string) []*Doc {
	if p.all == nil {
		return nil
	}
	docs := p.all[name]
	delete(p.all, name)
	return docs
}

func (p *pins) Walk(fn PinsWalkFunc) {
	if p.all == nil {
		return
	}
	for k, v := range p.all {
		fn(k, v)
	}
}

func (p *pins) Describe() {
	fmt.Println("Pins size", len(p.all))
	for k, _ := range p.all {
		fmt.Println("\t", k)
	}
}
