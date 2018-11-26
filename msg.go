package phly

const (
	WhatPins     = "pins"
	WhatStop     = "stop"
	whatFlush    = "flush"
	whatValidate = "validate"
)

// Msg is an abstract node message.
type Msg struct {
	What    string
	Payload interface{}
}

func MsgFromPins(pins Pins) Msg {
	return Msg{WhatPins, pins}
}

func MsgFromStop(err error) Msg {
	return Msg{WhatStop, &StopPayload{err}}
}

type StopPayload struct {
	Err error
}
