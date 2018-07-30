package phly

import ()

// --------------------------------
// NODE-DESCR

// NodeDescr describes a node.
type NodeDescr struct {
	Id         string
	Name       string
	Purpose    string
	InputPins  []PinDescr
	OutputPins []PinDescr
}

func (n *NodeDescr) ClaString() string {
	str := n.Name + " (" + n.Id + ")."
	if n.Purpose != "" {
		str += "\n" + n.Purpose
	}
	for _, descr := range n.InputPins {
		str += ("\n\tin \"" + descr.Name + "\". " + descr.Purpose)
	}
	for _, descr := range n.OutputPins {
		str += ("\n\tout \"" + descr.Name + "\". " + descr.Purpose)
	}
	return str
}

// --------------------------------
// PIN-DESCR

type PinDescr struct {
	Name    string
	Purpose string
}
