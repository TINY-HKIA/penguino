package discord

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
)

type Context interface {
	Send(msg *Message) error
	SendContent(msg string) error
	// Bind(i any) error
}

type InteractionContext struct {
	req interactionCreate
}

func (ctx InteractionContext) Send(msg *Message) error {
	resp := InteractionResponse{
		Type: Msg,
		Data: msg,
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		slog.Error("encode error:" + err.Error())
	}
	r, err := http.Post(httpApiBaseUrl+"/interactions/"+ctx.req.ID+"/"+ctx.req.Token+"/callback", "application/json", buf)
	if err != nil {
		slog.Error(err.Error())
	}

	slog.Debug("interaction_callback_response", "status", r.Status)
	return nil
}
func (ctx InteractionContext) SendContent(msg string) error {
	resp := InteractionResponse{
		Type: Msg,
		Data: &Message{Content: msg},
	}

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		slog.Error("encode error:" + err.Error())
	}
	r, err := http.Post(httpApiBaseUrl+"/interactions/"+ctx.req.ID+"/"+ctx.req.Token+"/callback", "application/json", buf)
	if err != nil {
		slog.Error(err.Error())
	}
	slog.Debug("interaction_callback_response", "status", r.Status)
	return nil
}
