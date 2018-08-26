package phly

import (
	"math/rand"
	"mime"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

var (
	env = &environment{}

	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

	texttype = mime.TypeByExtension(".txt")
)

const (
	text_txtoutput = "out"
)

func init() {
	// Prepare environment
	env.phlibPaths = append(env.phlibPaths, factoryPhlibPath())
	RegisterVar("cpus", "Number of CPUs", strconv.Itoa(runtime.NumCPU()))
	RegisterVar("os", "OS name", runtime.GOOS)
	RegisterVar("rndu", "Random unipolar number (0 to 1)", rndUnipolar)
	RegisterVar("rndb", "Random bipolar number (-1 to 1)", rndBipolar)

	// Register factory nodes
	Register(&batch{})
	Register(&console{})
	Register(&files{})
	Register(&filewatch{})
	Register(&pipeline{})
	Register(&run{})
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
