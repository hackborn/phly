package phly

import (
	"github.com/micro-go/parse"
)

// --------------------------------
// DOC

type Doc struct {
	Header
	Pages []*Page
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
