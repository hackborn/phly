package phly

import (
	"math/rand"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

var (
	env = &environment{}

	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

func init() {
	// Prepare environment
	env.phlibPaths = append(env.phlibPaths, factoryPhlibPath())
	RegisterVar("cpus", "Number of CPUs", strconv.Itoa(runtime.NumCPU()))
	RegisterVar("rndu", "Random unipolar number (0 to 1)", rndUnipolar)
	RegisterVar("rndb", "Random bipolar number (-1 to 1)", rndBipolar)

	// Register factory nodes
	Register(&batch{})
	Register(&files{})
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

func rndUnipolar() string {
	v := rnd.Float64()
	return strconv.FormatFloat(v, 'f', -1, 64)
}

func rndBipolar() string {
	v := rnd.Float64()*2 - 1
	return strconv.FormatFloat(v, 'f', -1, 64)
}
