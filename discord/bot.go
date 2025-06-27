package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Bot struct {
	token string
}
type Config struct {
	Token string
}

func NewBot(token string) *Bot {
	return &Bot{
		token: token,
	}
}

func NewBotWithConfig(cfg Config) *Bot {
	return &Bot{
		token: cfg.Token,
	}
}

func (bot *Bot) Start() error {
	ctx := context.Background()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", httpApiBaseUrl, getBotGateway), nil)
	req.Header.Set("Authorization", "Bot "+bot.token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid bot token")
	}

	b, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	var botGateway botGatewayResp
	if err := json.Unmarshal(b, &botGateway); err != nil {
		return err
	}

	slog.Debug("bot gateway response", "url", botGateway.URL)

	c, _, err := websocket.Dial(ctx, botGateway.URL, nil)
	if err != nil {
		return err
	}
	defer c.CloseNow()

	for {
		var payload EventPayload
		if err := wsjson.Read(ctx, c, &payload); err != nil {
			var closeErr websocket.CloseError
			if errors.As(err, &closeErr) {
				slog.Info("server closed connection",
					"code", closeErr.Code,
					"reason", closeErr.Reason)
				break
			} else {
				log.Fatal(err)
			}
		}

		slog.Info("inc payload:",
			"type", payload.Type,
			"sequence", payload.SequenceNum,
			"op", payload.Op)

		switch payload.Op {
		case 10:
			var helloReceiveEvent HelloReceiveEvent
			json.Unmarshal(payload.Data, &helloReceiveEvent)
			slog.Info(EventHelloReceive, "heartbeat_interval", helloReceiveEvent.HeartbeatInterval)
		default:
			slog.Warn("unrecognized opcode", "op", payload.Op, "data", string(payload.Data))
		}
	}
	return nil
}
