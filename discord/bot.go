package discord

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

var (
	DefaultConfig = Config{
		Token: os.Getenv("BOT_TOKEN"),
	}
)

type Bot struct {
	conn *websocket.Conn
	cfg  Config
}
type Config struct {
	Token string
}

func NewBot() *Bot {
	return &Bot{
		cfg: DefaultConfig,
	}
}

func NewBotWithConfig(cfg Config) *Bot {
	return &Bot{
		cfg: cfg,
	}
}

func (bot *Bot) Start() error {
	ctx := context.Background()

	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", httpApiBaseUrl, getBotGateway), nil)
	req.Header.Set("Authorization", "Bot "+bot.cfg.Token)

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

	slog.Debug("botGatewayResponse", "url", botGateway.URL)

	c, _, err := websocket.Dial(ctx, botGateway.URL, nil)
	if err != nil {
		return err
	}
	defer c.CloseNow()
	bot.conn = c

	for {
		var payload eventPayload
		if err := wsjson.Read(ctx, c, &payload); err != nil {
			var closeErr websocket.CloseError
			if errors.As(err, &closeErr) {
				slog.Info("server closed connection",
					"code", closeErr.Code,
					"reason", closeErr.Reason)
			}
			return err
		}

		slog.Debug("received",
			"type", payload.Type,
			"sequence", payload.SequenceNum,
			"op", payload.Op)
		switch payload.Op {
		case OpcodeHeartbeat:
			sendEvent(bot, gatewayEvent[any]{Op: OpcodeHeartbeat}, ctx)
		case OpcodeHello:
			var helloEvent helloReceiveEvent
			json.Unmarshal(payload.Data, &helloEvent)
			slog.Debug("helloEvent", "heartbeatInterval", helloEvent.HeartbeatInterval)

			sleepTime := float32(helloEvent.HeartbeatInterval) * rand.Float32()
			time.Sleep(time.Duration(sleepTime) * time.Millisecond)

			go bot.heartbeatLoop(ctx, helloEvent.HeartbeatInterval)
		case OpcodeHeartbeatAck:
		default:
			slog.Warn("unrecognizedOpcode", "op", payload.Op, "data", string(payload.Data))
		}
	}
}

func (bot *Bot) heartbeatLoop(ctx context.Context, interval int) {

	for {
		sendEvent(bot, gatewayEvent[any]{Op: 1}, ctx)
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

func sendEvent[T any](b *Bot, event gatewayEvent[T], ctx context.Context) {
	wsjson.Write(ctx, b.conn, event)
	slog.Debug("sent", "op", event.Op)
}

