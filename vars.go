package phly

import (
	"sort"
)

// RegisterVar() registers a variable with the system, which becomes part
// of the help system. Multiple variables with the same name can be registered,
// since each package can have its own non-conflicting variables. You can supply
// an optional value and the var wil be including when replacing vars through the
// Environment, although only a single var with the same name can have a value.
func RegisterVar(name, descr string, optional_value interface{}) {
	vardescrs = append(vardescrs, varDescr{name, descr})
	vardescrs = sortedVars(vardescrs)
	if optional_value != nil {
		env.setVar(name, optional_value)
	}
}

var (
	vardescrs = make([]varDescr, 0, 0)
)

type varDescr struct {
	name  string
	descr string
}

// --------------------------------
// SORT

func sortedVars(in []varDescr) []varDescr {
	var vars []varDescr
	for _, v := range in {
		vars = append(vars, v)
	}
	sort.Sort(SortVars(vars))
	return vars
}

type SortVars []varDescr

func (s SortVars) Len() int {
	return len(s)
}
func (s SortVars) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}
func (s SortVars) Less(i, j int) bool {
	return s[i].name < s[j].name
}
