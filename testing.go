package phly

// Support functions for testing

import (
	"encoding/json"
)

// ----------------------------------------
// STRING PINS

func StringPinsEqual(a, b Pins) bool {
	if a == nil && b == nil {
		return true
	} else if a != nil && b == nil {
		return false
	} else if a == nil && b != nil {
		return false
	}
	equal := true
	found := make(map[string]struct{})
	a.WalkPins(func(aname string, adocs Docs) {
		bdocs := b.GetPin(aname)
		found[aname] = struct{}{}
		if len(adocs.Docs) != len(bdocs.Docs) {
			equal = false
		} else {
			for i, ad := range adocs.Docs {
				if !stringDocsEqual(ad, bdocs.Docs[i]) {
					equal = false
				}
			}
		}
	})
	b.WalkPins(func(bname string, bdocs Docs) {
		if _, ok := found[bname]; !ok {
			equal = false
		}
	})
	return equal
}

func stringDocsEqual(a, b *Doc) bool {
	if a == nil && b == nil {
		return true
	} else if a != nil && b == nil {
		return false
	} else if a == nil && b != nil {
		return false
	}
	aitems := a.StringItems()
	bitems := b.StringItems()
	if len(aitems) != len(bitems) {
		return false
	}
	for i, av := range aitems {
		if av != bitems[i] {
			return false
		}
	}
	return true
}

func StringPinsToJson(pins Pins) string {
	if pins == nil {
		return "nil"
	}
	type StringDoc struct {
		Items []string
	}

	type StringDocs struct {
		Pins map[string][]*StringDoc
	}

	sd := &StringDocs{Pins: make(map[string][]*StringDoc)}
	pins.WalkPins(func(name string, docs Docs) {
		var t1 []*StringDoc
		for _, d := range docs.Docs {
			t2 := &StringDoc{Items: d.StringItems()}
			t1 = append(t1, t2)
		}
		sd.Pins[name] = t1
	})
	data, err := json.Marshal(sd)
	if err != nil {
		return err.Error()
	}
	return string(data)
}
