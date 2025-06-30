package discord

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand/v2"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Bot struct {
	cfg      Config
	conn     *websocket.Conn //
	seq      int
	handlers map[string]HandlerFunc
}

func NewBot() *Bot {
	return &Bot{
		cfg:      DefaultConfig,
		handlers: DefaultHandlers,
	}
}

func NewBotWithConfig(cfg Config) *Bot {
	return &Bot{
		cfg:      cfg,
		handlers: DefaultHandlers,
	}
}

type Config struct {
	Token string
}

var DefaultConfig = Config{
	Token: os.Getenv("BOT_TOKEN"),
}

var DefaultHandlers = make(map[string]HandlerFunc)

type HandlerFunc func(ctx Context) error

func (bot *Bot) Handle(command string, f HandlerFunc) {
	bot.handlers[command] = f
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
	slog.Debug("bot_gateway", "url", botGateway.URL)

	c, _, err := websocket.Dial(ctx, botGateway.URL+"/?v=10&encoding=json", nil)
	if err != nil {
		return err
	}
	defer c.CloseNow()
	bot.conn = c

	return bot.receiveLoop(ctx)
}

func (bot *Bot) receiveLoop(ctx context.Context) error {
	for {
		var payload receiveEvent
		if err := wsjson.Read(ctx, bot.conn, &payload); err != nil {
			// var closeErr websocket.CloseError
			return err
		}

		slog.Debug("receive", "op", payload.Op)

		switch payload.Op {
		case OpcodeDispatch:
			slog.Debug("dispatch",
				"t", *payload.Type,
				"s", *payload.SequenceNum,
				"d", string(payload.Data))

			bot.seq = *payload.SequenceNum

			switch *payload.Type {
			case EventReady:
				slog.Info(penguinoStart)
			case EventInteractionCreate:
				var d interactionCreate
				if err := json.Unmarshal(payload.Data, &d); err != nil {
					return err
				}

				err := bot.handlers[d.Data.Name](InteractionContext{req: d})
				if err != nil {
					slog.Error(err.Error())
				}
			}
		case OpcodeHeartbeat:
			write(ctx, bot, sendEvent[any]{Op: OpcodeHeartbeat})
		case OpcodeHello:
			var helloEvent helloReceivePayload
			json.Unmarshal(payload.Data, &helloEvent)
			slog.Debug("heartbeat", "interval", helloEvent.HeartbeatInterval)
			go bot.heartbeatLoop(ctx, helloEvent.HeartbeatInterval)
			write(ctx, bot, sendEvent[any]{
				Op: OpcodeIdentify,
				Data: map[string]any{
					"token":   bot.cfg.Token,
					"intents": 513,
					"properties": map[string]any{
						"os": runtime.GOOS,
					},
					"presence": map[string]any{
						"activities": []map[string]any{
							{
								"name": "HKIA!!!!!",
								"type": 0,
							},
						},
						"status": status,
					},
				}})
		case OpcodeHeartbeatAck:
		default:
			slog.Warn("unknown opcode", "op", payload.Op, "data", string(payload.Data))
		}
	}
}

func (bot *Bot) heartbeatLoop(ctx context.Context, interval int) {
	// wait for (heartbeat_interval * jitter) milliseconds before starting cycle
	sleepTime := float32(interval) * rand.Float32()
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)

	for {
		write(ctx, bot, sendEvent[any]{Op: 1})
		time.Sleep(time.Duration(interval) * time.Millisecond)
	}
}

func write[T any](ctx context.Context, b *Bot, event sendEvent[T]) {
	wsjson.Write(ctx, b.conn, event)
	slog.Debug("send", "op", event.Op, "s", b.seq)
}
