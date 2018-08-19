package phly

import (
	"bytes"
	"io"
)

// --------------------------------
// NAMER

// namer provides a name
type namer interface {
	Name() string
}

// --------------------------------
// NAMED-READER

// namedReader wraps a reader in a named object.
type namedReader struct {
	name   string
	reader io.Reader
}

func newNamedReader(name string, data []byte) io.Reader {
	return &namedReader{name, bytes.NewReader(data)}
}

func (n *namedReader) Name() string {
	return n.name
}

func (n *namedReader) Read(p []byte) (int, error) {
	return n.reader.Read(p)
}
