package phly

import (
	"os"
	"path/filepath"
	"runtime"
	"strconv"
)

var (
	env = &environment{}
)

func init() {
	// Prepare environment
	env.phlibPaths = append(env.phlibPaths, factoryPhlibPath())
	RegisterVar("cpus", "Number of CPUs", strconv.Itoa(runtime.NumCPU()))

	// Register factory nodes
	Register(&batch{})
	Register(&pipeline{})
	Register(&text{})
}

func factoryPhlibPath() string {
	ex, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Join(filepath.Dir(ex), "phlib")
}
