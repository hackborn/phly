package phly

import (
//	"fmt"
)

// ----------------------------------------
// PINS

// Pins describes a collection of documents attached to named pins.
// It is meant for clients that are consuming pins.
type Pins interface {
	GetPin(name string) Docs
	WalkPins(fn PinsWalkFunc)
}

// ----------------------------------------
// PIN FUNCTIONS

// PinsWalkFunc iterates over each pin in a Pins collection.
type PinsWalkFunc func(name string, docs Docs)

// ItemFunc is a callback for a single item in a pin doc.
type ItemFunc func(channel string, doc *Doc, index int, item interface{})

// StringItemFunc is a callback for a single string item in a pin doc.
type StringItemFunc func(channel string, doc *Doc, index int, item string)

// ----------------------------------------
// PIN ITERATION

// WalkItems iterates over each item on the channel
func WalkItems(pins Pins, channel string, fn ItemFunc) {
	for _, doc := range pins.GetPin(channel).Docs {
		for idx, item := range doc.Items {
			fn(channel, doc, idx, item)
		}
	}
}

// WalkStringItems iterates over each string item on the channel
func WalkStringItems(pins Pins, channel string, fn StringItemFunc) {
	for _, doc := range pins.GetPin(channel).Docs {
		for idx, _item := range doc.Items {
			if item, ok := _item.(string); ok {
				fn(channel, doc, idx, item)
			}
		}
	}
}

// ----------------------------------------
// PINS

type pins struct {
	all map[string]*Docs
}

func (p *pins) GetPin(name string) Docs {
	if p.all == nil {
		return Docs{}
	}
	docs := p.all[name]
	if docs != nil {
		return *docs
	}
	return Docs{}
}

func (p *pins) WalkPins(fn PinsWalkFunc) {
	for k, v := range p.all {
		if v != nil {
			fn(k, *v)
		}
	}
}

func (p *pins) add(name string, item *Doc) {
	if item == nil {
		return
	}
	if p.all == nil {
		p.all = make(map[string]*Docs)
	}
	docs := p.all[name]
	if docs == nil {
		docs = &Docs{}
	}
	docs.Docs = append(docs.Docs, item)
	p.all[name] = docs
}

func (p *pins) addDocs(name string, docs *Docs) {
	if docs == nil {
		return
	}
	for _, doc := range docs.Docs {
		p.add(name, doc)
	}
}
