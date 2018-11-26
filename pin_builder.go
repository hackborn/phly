package phly

import (
	"errors"
)

// ----------------------------------------
// PIN-BUILDER

type PinBuilder struct {
	build pins
}

func (b PinBuilder) Pins() Pins {
	return &b.build
}

func (b PinBuilder) Add(name string, item *Doc) PinBuilder {
	b.build.add(name, item)
	return b
}

// ----------------------------------------
// BUILD-PINS STREAM
// Create pins via an esoteric but convenience series of commands.

type BuildPinsCmd int

const (
	pbsEmpty BuildPinsCmd = iota // Internal empty cmd
	PbsChan                      // The following command must be a string. It will be used to
	// start a new pin with the given channel name.
	PbsDoc // Start a new doc. Can be ommitted after starting a channel. Only necessary
	// when you want multiple docs on the same pin.
)

// BuildPins() builds a new Pins object according to the command
// stream. See the tokens for the rules. Any cmd that is not a token or made
// use of by the last token becomes a new item in the current doc.
// Example. The first command defaults to being the channel (so it must be a string).
//		BuildPins("input", "item1", "item2")
// Example. You can use the channel command to create a new channel at any point.
//		BuildPins(PbsChannel, "input", "item1", "item2")
// Example. You can use the doc command to create a new doc at any point.
//		BuildPins(PbsChannel, "input", "doc1_item1", PbsDoc, "doc2_item1")
func BuildPins(cmds ...interface{}) (Pins, error) {
	channels := make(map[string]*Docs)
	var cur *Docs
	// The first command always needs to be a channel, though actually
	// specifying the channel command is optional
	lastcmd := PbsChan
	for _, _cmd := range cmds {
		switch cmd := _cmd.(type) {
		case BuildPinsCmd:
			switch cmd {
			case PbsChan:
				lastcmd = cmd
			case PbsDoc:
				lastcmd = pbsEmpty
				if cur == nil {
					return nil, errors.New("Build on nil channel")
				}
				cur.appendDoc(nil)
			}
		case string:
			if lastcmd == PbsChan {
				cur = &Docs{}
				channels[cmd] = cur
				lastcmd = pbsEmpty
			} else {
				if lastcmd != pbsEmpty {
					return nil, errors.New("Build on invalid command stream")
				}
				pbsAppendItem(cur, cmd)
			}
		case *Docs:
			if cur == nil {
				return nil, errors.New("Build on nil channel")
			}
			for _, doc := range cmd.Docs {
				cur.appendDoc(doc)
			}
		default:
			if lastcmd != pbsEmpty {
				return nil, errors.New("Build on invalid command stream")
			}
			pbsAppendItem(cur, cmd)
		}
	}
	return &pins{channels}, nil
}

// MustBuildPins() is identical to BuildPins but it will panic on an
// error. Intended only for testing
func MustBuildPins(cmds ...interface{}) Pins {
	pins, err := BuildPins(cmds...)
	if err != nil {
		panic(err)
	}
	return pins
}

func pbsAppendItem(docs *Docs, item interface{}) {
	if docs == nil {
		panic("Build on nil channel")
	}
	if len(docs.Docs) < 1 {
		docs.appendDoc(nil)
	}
	docs.Docs[len(docs.Docs)-1].AppendItem(item)
}
