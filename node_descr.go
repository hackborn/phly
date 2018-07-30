package phly

import ()

// --------------------------------
// NODE-DESCR

// NodeDescr describes a node.
type NodeDescr struct {
	Id         string
	Name       string
	Purpose    string
	Cfgs       []CfgDescr
	InputPins  []PinDescr
	OutputPins []PinDescr
}

func (n *NodeDescr) ClaString() string {
	str := n.Name + " (" + n.Id + ")."
	if n.Purpose != "" {
		str += "\n" + n.Purpose
	}
	for _, descr := range n.Cfgs {
		str += ("\n\tcfg \"" + descr.Name + "\". " + descr.Purpose)
	}
	for _, descr := range n.InputPins {
		str += ("\n\tinput \"" + descr.Name + "\". " + descr.Purpose)
	}
	for _, descr := range n.OutputPins {
		str += ("\n\toutput \"" + descr.Name + "\". " + descr.Purpose)
	}
	return str
}

func (n *NodeDescr) MarkdownString() string {
	str := "* **" + n.Name + "** (" + n.Id + ")."
	if n.Purpose != "" {
		str += " " + n.Purpose
	}
	for _, descr := range n.Cfgs {
		str += ("\n    * cfg **" + descr.Name + "**. " + descr.Purpose)
	}
	for _, descr := range n.InputPins {
		str += ("\n    * input **" + descr.Name + "**. " + descr.Purpose)
	}
	for _, descr := range n.OutputPins {
		str += ("\n    * output **" + descr.Name + "**. " + descr.Purpose)
	}
	return str
}

// --------------------------------
// CFG-DESCR

type CfgDescr struct {
	Name    string
	Purpose string
}

// --------------------------------
// PIN-DESCR

type PinDescr struct {
	Name    string
	Purpose string
}
