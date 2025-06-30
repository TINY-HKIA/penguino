package discord

import (
	"encoding/json"
)

const (
	OpcodeDispatch     = 0
	OpcodeHeartbeat    = 1
	OpcodeIdentify     = 2
	OpcodeResume       = 6
	OpcodeReconnect    = 7
	OpcodeHello        = 10
	OpcodeHeartbeatAck = 11

	EventReady             = "READY"
	EventInteractionCreate = "INTERACTION_CREATE"
	EventResumed           = "RESUMED"
)

// Incoming gateway event
type receiveEvent struct {
	Type        *string         `json:"t"` // unknown atm
	SequenceNum *int            `json:"s"` // nullable int
	Op          int             `json:"op"`
	Data        json.RawMessage `json:"d"`
}

// Outgoing gateway event
type sendEvent[T any] struct {
	Op   int `json:"op"`
	Data T   `json:"d"`
}

type helloReceivePayload struct {
	HeartbeatInterval int `json:"heartbeat_interval"`
}

type resumeSendPayload struct {
	SessionID string `json:"session_id"`
	Token     string `json:"token"`
	Seq       int    `json:"seq"`
}
