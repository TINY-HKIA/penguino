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
	Cfg Config

	conn       *websocket.Conn //
	handlers   map[string]HandlerFunc
	gatewayUrl string
	seq        int
	isResumed  bool
	sessionId  string
}

func NewBot() *Bot {
	return &Bot{
		Cfg:      DefaultConfig,
		handlers: DefaultHandlers,
	}
}

func NewBotWithConfig(cfg Config) *Bot {
	return &Bot{
		Cfg:      cfg,
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

func (bot *Bot) HandleCommand(name string, f HandlerFunc) {
	bot.handlers[name] = f
}

func (bot *Bot) Start() error {
	ctx := context.Background()
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", httpApiBaseUrl, getBotGateway), nil)
	req.Header.Set("Authorization", "Bot "+bot.Cfg.Token)

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
	bot.gatewayUrl = botGateway.URL

	c, _, err := websocket.Dial(ctx, bot.gatewayUrl+"/?v=10&encoding=json", nil)
	if err != nil {
		return err
	}
	defer func() {
		if c != nil {
			c.CloseNow()
		}
	}()
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
			bot.seq = *payload.SequenceNum
			slog.Debug("dispatch",
				"t", *payload.Type,
				"s", *payload.SequenceNum)
			switch *payload.Type {
			case EventReady:
				var ready struct {
					ResumeUrl string `json:"resume_gateway_url"`
					SessionId string `json:"session_id"`
				}
				json.Unmarshal(payload.Data, &ready)
				bot.gatewayUrl = ready.ResumeUrl
				bot.sessionId = ready.SessionId
				slog.Info(penguinoStart)

			case EventInteractionCreate:
				// slog.Debug("dispatch",
				// 	"t", *payload.Type,
				// 	"s", *payload.SequenceNum,
				// 	"d", string(payload.Data))

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
			if !bot.isResumed {
				go bot.heartbeatLoop(ctx, helloEvent.HeartbeatInterval)
				if err := bot.identify(ctx); err != nil {
					return err
				}
			}

		case OpcodeHeartbeatAck:
		case OpcodeReconnect:
			bot.conn.CloseNow()
			c, _, err := websocket.Dial(ctx, bot.gatewayUrl+"/?v=10&encoding=json", nil)
			if err != nil {
				return err
			}
			bot.isResumed = true
			bot.conn = c
			write(ctx, bot, sendEvent[resumeSendPayload]{Op: OpcodeResume, Data: resumeSendPayload{
				SessionID: bot.sessionId,
				Token:     bot.Cfg.Token,
				Seq:       bot.seq,
			}})
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

func (b *Bot) identify(ctx context.Context) error {
	identify := sendEvent[any]{
		Op: OpcodeIdentify,
		Data: map[string]any{
			"token":   b.Cfg.Token,
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
		}}

	return write(ctx, b, identify)
}

func write[T any](ctx context.Context, b *Bot, event sendEvent[T]) error {
	slog.Debug("send", "op", event.Op, "s", b.seq)
	return wsjson.Write(ctx, b.conn, event)
}
