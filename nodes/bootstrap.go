package phly_nodes

import (
	"github.com/hackborn/phly"
	"mime"
)

var (
	texttype = mime.TypeByExtension(".txt")
)

func init() {
	// Register nodes
	//	phly.Register(&batch{})
	//	phly.Register(&console{})
	phly.Register(&files{})
	//	phly.Register(&filewatch{})
	phly.Register(&run{})
}
