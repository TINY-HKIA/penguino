package discord

import "encoding/json"

const (
	EventHelloReceive string = "HelloReceiveEvent"
)

type EventPayload struct {
	Type        any             `json:"t"` // unknown
	SequenceNum *int            `json:"s"` // nullable int
	Op          int             `json:"op"`
	Data        json.RawMessage `json:"d"`
}

type HelloReceiveEvent struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}
