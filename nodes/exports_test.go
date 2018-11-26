package phly_nodes

import (
	"github.com/hackborn/phly"
)

func New_run() phly.Node {
	return &run{}
}

const (
	Run_cmdinput  = run_cmdinput
	Run_clainput  = run_clainput
	Run_output    = run_output
	Run_erroutput = run_erroutput
)
