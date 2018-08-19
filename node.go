package phly

import (
	"path/filepath"
)

// --------------------------------
// NODE

// Node performs abstract document processing.
// Nodes that need to perform clean up should implement io.Closer.
type Node interface {
	Describe() NodeDescr
	Run(args RunArgs, input Pins, sender PinSender) (Flow, error)
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
// PIN-SENDER

// PinSender sends pins from a node to a destination.
type PinSender interface {
	SendPins(Node, Pins)
	SendFinished(Node, error)
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
	Fields     map[string]interface{}
	workingdir string            // All relative file paths will use this as the root.
	cla        map[string]string // Command line arguments
	stop       chan struct{}
}

func (r *RunArgs) copy() RunArgs {
	fields := make(map[string]interface{})
	return RunArgs{r.Env, fields, r.workingdir, r.cla, r.stop}
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
	abs, err := filepath.Abs(filepath.Join(r.workingdir, rel))
	if err != nil {
		return rel
	}
	return abs
}
