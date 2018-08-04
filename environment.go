package phly

import (
	"bytes"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Environment provides access to the system environment.
type Environment interface {
	// FindFile() answers the full path to the phlyp by searching paths for name.
	FindFile(name string) string
	// FindReader() answers a reader for the given name.
	FindReader(name string) io.Reader
}

// Environment stores the current phly environment.
type environment struct {
	phlibPaths []string // A list of all registered library locations.
	phlypCache map[string][]byte
}

// FindFile() answers the full path to the phlyp by searching paths for name.
func (e *environment) FindFile(name string) string {
	for _, dir := range e.phlibPaths {
		p := filepath.Join(dir, name)
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}
	return ""
}

// FindReader() answers a reader to the path. A cache is used so reuse is efficient.
func (e *environment) FindReader(name string) io.Reader {
	resolved := e.FindFile(name)
	if name == "" {
		return nil
	}
	if e.phlypCache == nil {
		e.phlypCache = make(map[string][]byte)
	}
	if data, ok := e.phlypCache[resolved]; ok {
		return bytes.NewReader(data)
	}
	data, err := ioutil.ReadFile(resolved)
	if err != nil {
		return nil
	}
	e.phlypCache[resolved] = data
	return bytes.NewReader(data)
}
