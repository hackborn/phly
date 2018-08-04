package phly

import (
	"bytes"
	"github.com/micro-go/parse"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

// Environment provides access to the system environment.
type Environment interface {
	// Answer the system var with the given name.
	Var(name string) string
	// FindFile() answers the full path to the phlyp by searching paths for name.
	FindFile(name string) string
	// FindReader() answers a reader for the given name.
	FindReader(name string) io.Reader

	// Utility for replacing strings with a collection of my vars and supplied pairs.
	ReplaceVars(s string, pairs ...interface{}) string
}

// Environment stores the current phly environment.
type environment struct {
	vars        map[string]string
	replaceVars []interface{}
	phlibPaths  []string // A list of all registered library locations.
	phlypCache  map[string][]byte
}

func (e *environment) Var(name string) string {
	if e.vars == nil {
		return ""
	}
	return e.vars[name]
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

func (e *environment) ReplaceVars(s string, pairs ...interface{}) string {
	if len(e.replaceVars) > 0 {
		s = parse.ReplacePairs(s, e.replaceVars...)
	}
	if len(pairs) > 0 {
		s = parse.ReplacePairs(s, pairs...)
	}
	return s
}

func (e *environment) setVar(name, value string) {
	if e.vars == nil {
		e.vars = make(map[string]string)
	}
	e.vars[name] = value
	e.replaceVars = append(e.replaceVars, "${"+name+"}")
	e.replaceVars = append(e.replaceVars, value)
}