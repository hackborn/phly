package phly

import (
	"github.com/micro-go/parse"
)

// ----------------------------------------
// DOC

// Doc describes a single abstract document. It includes
// a user-defined header, an optional content type, and optional pages.
type Doc struct {
	Header
	MimeType string
	Items    []interface{}
}

// Create a new doc on string items
func NewStringDoc(n ...string) *Doc {
	doc := &Doc{}
	for _, i := range n {
		doc.Items = append(doc.Items, i)
	}
	return doc
}

func (d *Doc) AppendItem(item interface{}) {
	d.Items = append(d.Items, item)
}

func (d *Doc) AllItems() []interface{} {
	return d.Items
}

func (d *Doc) StringItems() []string {
	var dst []string
	for _, _s := range d.Items {
		if s, ok := _s.(string); ok {
			dst = append(dst, s)
		}
	}
	return dst
}

func (d *Doc) AllItem(index int) interface{} {
	if len(d.Items) <= index {
		return nil
	}
	return d.Items[index]
}

func (d *Doc) StringItem(index int) string {
	if len(d.Items) <= index {
		return ""
	}
	if s, ok := d.Items[index].(string); ok {
		return s
	}
	return ""
}

// ----------------------------------------
// DOCS

// Docs is a wrapper on a slice of Doc structs, used
// to add convenience management functions.
type Docs struct {
	Docs []*Doc
}

func NewDocs(docs ...*Doc) *Docs {
	ans := &Docs{}
	for _, d := range docs {
		ans.Docs = append(ans.Docs, d)
	}
	return ans
}

func (d *Docs) appendDoc(doc *Doc) {
	if doc == nil {
		doc = &Doc{}
	}
	d.Docs = append(d.Docs, doc)
}

func (d Docs) AllItems() []interface{} {
	var items []interface{}

	// optimizations
	if len(d.Docs) < 1 {
		return items
	}
	if len(d.Docs) == 1 {
		return d.Docs[0].Items
	}

	for _, d := range d.Docs {
		if len(d.Items) > 0 {
			items = append(items, d.Items...)
		}
	}
	return items
}

func (d Docs) StringItems() []string {
	var dst []string

	for _, d := range d.Docs {
		items := d.StringItems()
		if len(items) > 0 {
			dst = append(dst, items...)
		}
	}
	return dst
}

func (d Docs) AllItem(index int) interface{} {
	if len(d.Docs) < 1 {
		return nil
	}
	for _, d := range d.Docs {
		if index < len(d.Items) {
			return d.Items[index]
		}
		index -= len(d.Items)
	}
	return nil
}

func (d Docs) StringItem(index int) string {
	_s := d.AllItem(index)
	if s, ok := _s.(string); ok {
		return s
	}
	return ""
}

// ----------------------------------------
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

/*
func NewDocOnStringItems(n ...string) *Doc {
	doc := &Doc{}
	page := doc.NewPage("")
	for _, i := range n {
		page.Items = append(page.Items, i)
	}
	return doc
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

func (d *Doc) GetItems() []interface{} {
	if len(d.Pages) < 1 {
		return nil
	}
	return d.Pages[0].Items
}

func (d *Doc) GetStringItems() []string {
	src := d.GetItems()
	var dst []string
	for _, _s := range src {
		if s, ok := _s.(string); ok {
			dst = append(dst, s)
		}
	}
	return dst
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
// MISC

func copyDocs(src []*Doc) []*Doc {
	// XXX Obviously a big hole, but we're not using yet,
	// and not sure what the proper solution is here, there
	// doesn't look like a performant way to copy interface{}.
	return src
}
*/
