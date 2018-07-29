package phly

// Pins describes a collection of documents attached to pins.
type Pins interface {
	Get(name string) []*Doc
	Add(name string, doc *Doc)
}

type pins struct {
	all map[string][]*Doc
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
