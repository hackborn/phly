package phly

import (
	"path/filepath"
)

// ----------------------------------------
// INSTANTIATE-ARGS

// InstantiateArgs provides information during the instantiation phase.
type InstantiateArgs struct {
	Env Environment
}

// ----------------------------------------
// PROCESS-ARGS

// ProcessArgs provides arguments to the node during processing.
type ProcessArgs struct {
	//	Fields     map[string]interface{} // Unused?

	env        Environment
	dryRun     bool
	workingdir string            // All relative file paths will use this as the root.
	cla        map[string]string // Command line arguments
	stop       chan struct{}
}

func (r *ProcessArgs) Env() Environment {
	return r.env
}

// ClaValue() answers the command line argument value for the given name.
func (r *ProcessArgs) ClaValue(name string) string {
	if name == "" || r.cla == nil {
		return ""
	}
	return r.cla[name]
}

// Filename() answers an absolute filename for the supplied filename.
// Absolute filenames are returned as-is. Relative filenames are
// made relative to the cfg that generated the pipeline.
func (r *ProcessArgs) Filename(rel string) string {
	if r.workingdir == "" {
		return rel
	}
	if filepath.IsAbs(rel) {
		return rel
	}
	abs, err := filepath.Abs(filepath.Join(r.workingdir, rel))
	if err != nil {
		return rel
	}
	return abs
}

func (r *ProcessArgs) copy() *ProcessArgs {
	//	fields := make(map[string]interface{})
	return &ProcessArgs{r.env, r.dryRun, r.workingdir, r.cla, r.stop}
}

// ----------------------------------------
// STOPPED-ARGS

// StoppedArgs provides information as nodes are stopping.
type StoppedArgs struct {
}
