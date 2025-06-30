package discord

import "encoding/json"

const (
	OpcodeDispatch     = 0
	OpcodeHeartbeat    = 1
	OpcodeIdentify     = 2
	OpcodeReconnect    = 7
	OpcodeHello        = 10
	OpcodeHeartbeatAck = 11
)

const (
	EventReady             = "READY"
	EventInteractionCreate = "INTERACTION_CREATE"
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

var defaultIdentifyPayload = map[string]any{
	"op": 2,
	"d": map[string]any{
		"token":   "your_token_here",
		"intents": 8,
		"properties": map[string]any{
			"os": "linux",
		},
		"presence": map[string]any{
			"activities": []map[string]any{
				{
					"name": "HKIA!!!!!",
					"type": 0,
				},
			},
			"status": "dnd",
		},
	},
}
