package discord

import (
	"bytes"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
)

type Context interface {
	Respond(resp InteractionResponse) error
	Bind(i any) error
}

type InteractionContext struct {
	req interactionCreate
}

func (ctx InteractionContext) Bind(i any) error {
	return nil
}

func (ctx InteractionContext) Respond(resp InteractionResponse) error {

	buf := new(bytes.Buffer)
	if err := json.NewEncoder(buf).Encode(resp); err != nil {
		slog.Error("encode error:" + err.Error())
	}
	r, err := http.Post(httpApiBaseUrl+"/interactions/"+ctx.req.ID+"/"+ctx.req.Token+"/callback", "application/json", buf)
	if err != nil {
		slog.Error(err.Error())
	}
	defer r.Body.Close()

	b, err := io.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		return err
	}
	slog.Debug("interaction callback response", "status", r.Status, "body", string(b))
	return nil
}
