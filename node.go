package phly

import (
	"path/filepath"
)

// --------------------------------
// NODE

// Node performs abstract document processing.
type Node interface {
	Describe() NodeDescr
	Run(args RunArgs, input, output Pins) error
}

// --------------------------------
// NODE-FACTORY

// NodeFactory instantiates a node. All nodes should have their
// factories installed in the bootstrap.
type NodeFactory interface {
	Describe() NodeDescr
	Instantiate(tree interface{}) (Node, error)
}

// --------------------------------
// RUN-ARGS

// RunArgs provides arguments to the node during the run.
type RunArgs struct {
	Cla        []string // Command line arguments
	WorkingDir string   // All relative file paths will use this as the root.
	Fields     map[string]interface{}
}

// Filename() answers an absolute filename for the supplied filename.
// Absolute filenames are returned as-is. Relative filenames are
// made relative to the cfg that generated the pipeline.
func (r *RunArgs) Filename(rel string) string {
	if filepath.IsAbs(rel) {
		return rel
	}
	abs, err := filepath.Abs(filepath.Join(r.WorkingDir, rel))
	if err != nil {
		return rel
	}
	return abs
}
