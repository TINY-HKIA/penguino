package discord

import "encoding/json"

// Gateway Opcodes
const (
	OpcodeHeartbeat    = 1
	OpcodeReconnect    = 7
	OpcodeHello        = 10
	OpcodeHeartbeatAck = 11
)

// Incoming gateway event
type eventPayload struct {
	Type        any             `json:"t"` // unknown atm
	SequenceNum *int            `json:"s"` // nullable int
	Op          int             `json:"op"`
	Data        json.RawMessage `json:"d"`
}

// Outgoing gateway event 
type gatewayEvent[T any] struct {
	Op   int `json:"op"`
	Data T   `json:"d"`
}

type helloReceiveEvent struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}
