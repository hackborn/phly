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
	Instantiate(args InstantiateArgs, tree interface{}) (Node, error)
}

// --------------------------------
// INSTANTIATE-ARGS

// InstantiateArgs provides information during the instantiation phase.
type InstantiateArgs struct {
	Env Environment
}

// --------------------------------
// RUN-ARGS

// RunArgs provides arguments to the node during the run.
type RunArgs struct {
	Env        Environment
	WorkingDir string            // All relative file paths will use this as the root.
	cla        map[string]string // Command line arguments
	Fields     map[string]interface{}
	nodename   string // The name of the node currently using this run.
}

func (r *RunArgs) copy() RunArgs {
	fields := make(map[string]interface{})
	return RunArgs{r.Env, r.WorkingDir, r.cla, fields, r.nodename}
}

// ClaValue() answers the command line argument value for the given name.
func (r *RunArgs) ClaValue(name string) string {
	if name == "" || r.cla == nil {
		return ""
	}
	return r.cla[name]
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
