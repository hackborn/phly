package phly

import (
	"github.com/micro-go/parse"
)

// --------------------------------
// DOC

// Doc describes a single abstract document. It includes
// a user-defined header, an optional content type, and optional pages.
type Doc struct {
	Header
	MimeType string
	Pages    []*Page
}

func (d *Doc) NewPage(name string) *Page {
	page := &Page{Name: name}
	d.Pages = append(d.Pages, page)
	return page
}

func (d *Doc) AddPage(page *Page) *Page {
	d.Pages = append(d.Pages, page)
	return page
}

func (d *Doc) FindPage(name string) *Page {
	if d.Pages == nil {
		return nil
	}
	for _, t := range d.Pages {
		if t.Name == name {
			return t
		}
	}
	return nil
}

func (d *Doc) RemovePage(name string) bool {
	if d.Pages == nil {
		return false
	}
	for i, t := range d.Pages {
		if t.Name == name {
			d.Pages = append(d.Pages[:i], d.Pages[i+1:]...)
			return true
		}
	}
	return false
}

// --------------------------------
// PAGE

type Page struct {
	Header
	Name  string
	Items []interface{}
}

func (p *Page) Copy() *Page {
	dst := &Page{Name: p.Name}
	dst.Header.Values = p.Header.Values
	for _, i := range p.Items {
		dst.Items = append(dst.Items, i)
	}
	return dst
}

func (p *Page) AddItem(v interface{}) *Page {
	p.Items = append(p.Items, v)
	return p
}

// --------------------------------
// HEADER

type Header struct {
	Values interface{}
}

func (h *Header) SetHeader(values interface{}) {
	h.Values = values
}

func (h *Header) GetInt(path string) (int, bool) {
	if h.Values == nil {
		return 0, false
	}
	return parse.FindTreeInt(path, h.Values)
}

func (h *Header) GetString(path string) (string, bool) {
	if h.Values == nil {
		return "", false
	}
	return parse.FindTreeString(path, h.Values)
}

func (h *Header) SetInt(path string, value int) error {
	v, err := parse.SetTreeInt(path, value, h.Values)
	if err == nil {
		h.Values = v
	}
	return err
}

func (h *Header) SetString(path, value string) error {
	v, err := parse.SetTreeString(path, value, h.Values)
	if err == nil {
		h.Values = v
	}
	return err
}

// --------------------------------
// MISC

func copyDocs(src []*Doc) []*Doc {
	// XXX Obviously a big hole, but we're not using yet,
	// and not sure what the proper solution is here, there
	// doesn't look like a performant way to copy interface{}.
	return src
}
