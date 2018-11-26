package phly

import ()

// ----------------------------------------
// ITEMS

// Items provides common behaviour on an
// abstract collection.
type Items interface {
	// Get collections in various formats
	AllItems() []interface{}
	StringItems() []string

	// Get at index in various formats
	AllItem(index int) interface{}
	StringItem(index int) string
}
