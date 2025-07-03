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
	Cfg     Config
	Context context.Context

	handlers   map[string]HandlerFunc
	gatewayUrl string

	seq         int
	isReconnect bool
	sessionId   string
}

func NewBot(ctx context.Context) *Bot {
	return &Bot{
		Cfg:      DefaultConfig,
		Context:  ctx,
		handlers: DefaultHandlers,
	}
}

func NewBotWithConfig(cfg Config, ctx context.Context) *Bot {
	return &Bot{
		Cfg:      cfg,
		Context:  ctx,
		handlers: DefaultHandlers,
	}
}

type Config struct {
	Token string
}

type session struct {
	Conn    *websocket.Conn
	Context context.Context

	eventch chan sendEvent[any]
}

func (b *Bot) Connect() (*session, error) {
	if !b.isReconnect {
		if err := b.initBotGateway(); err != nil {
			return nil, err
		}
	}

	c, _, err := websocket.Dial(b.Context, b.gatewayUrl+"/?v=10&encoding=json", nil)
	if err != nil {
		return nil, err
	}

	sesh := &session{
		Conn:    c,
		eventch: make(chan sendEvent[any]),
	}

	return sesh, nil
}

var DefaultConfig = Config{
	Token: os.Getenv("BOT_TOKEN"),
}

var DefaultHandlers = make(map[string]HandlerFunc)

type HandlerFunc func(ctx Context) error

func (b *Bot) HandleCommand(name string, f HandlerFunc) {
	b.handlers[name] = f
}

func (b *Bot) Start() error {
	for {
		select {
		case <-b.Context.Done():
			return b.Context.Err()
		default:
			sesh, err := b.Connect()
			if err != nil {
				return err
			}
			ctx, cancel := context.WithCancel(b.Context)
			sesh.Context = ctx

			go sesh.handleWrites()
			slog.Warn(b.readLoop(sesh, ctx, cancel).Error())
		}
	}
}

func (s *session) handleWrites() {
	for {
		select {
		case <-s.Context.Done():
			slog.Info("writer closing")
			return
		case event, ok := <-s.eventch:
			if !ok {
				slog.Warn("attempted read on closed eventch")
			}

			slog.Debug("send", "op", event.Op)
			if err := wsjson.Write(s.Context, s.Conn, event); err != nil {
				slog.Error("write err", "err", err.Error())
			}
		}
	}

}

func (b *Bot) readLoop(sesh *session, ctx context.Context, cancel context.CancelFunc) error {
	defer sesh.Conn.CloseNow()
	for {
		var payload receiveEvent
		if err := wsjson.Read(ctx, sesh.Conn, &payload); err != nil {
			cancel()
			slog.Debug("stopping readLoop", "sessionId", b.sessionId)
			return err
		}

		slog.Debug("receive", "op", payload.Op)

		switch payload.Op {
		case OpcodeDispatch:
			b.seq = *payload.SequenceNum
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
				b.gatewayUrl = ready.ResumeUrl
				b.sessionId = ready.SessionId
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

				err := b.handlers[d.Data.Name](InteractionContext{req: d})
				if err != nil {
					slog.Error(err.Error())
				}
			}
		case OpcodeHeartbeat:
			sesh.eventch <- sendEvent[any]{Op: OpcodeHeartbeat}
		case OpcodeHello:
			var helloEvent helloReceivePayload
			json.Unmarshal(payload.Data, &helloEvent)
			slog.Debug("heartbeat", "interval", helloEvent.HeartbeatInterval)
			go sesh.heartbeatLoop(ctx, helloEvent.HeartbeatInterval)
			if b.isReconnect {
				sesh.eventch <- sendEvent[any]{Op: OpcodeResume, Data: resumeSendPayload{
					SessionID: b.sessionId,
					Token:     b.Cfg.Token,
					Seq:       b.seq,
				}}

			} else {
				b.identify(sesh)
			}

		case OpcodeHeartbeatAck:
		case OpcodeReconnect:
			slog.Warn("reconnecting")
			cancel()
			b.isReconnect = true
		default:
			slog.Warn("unknown opcode", "op", payload.Op, "data", string(payload.Data))
		}
	}
}

func (sesh *session) heartbeatLoop(ctx context.Context, interval int) {
	// wait for (heartbeat_interval * jitter) milliseconds before starting cycle
	sleepTime := float32(interval) * rand.Float32()
	time.Sleep(time.Duration(sleepTime) * time.Millisecond)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			sesh.eventch <- sendEvent[any]{Op: 1}
			time.Sleep(time.Duration(interval) * time.Millisecond)
		}
	}
}

func (b *Bot) identify(sesh *session) {
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

	sesh.eventch <- identify
}

func (b *Bot) initBotGateway() error {
	req, _ := http.NewRequest(http.MethodGet, fmt.Sprintf("%s%s", httpApiBaseUrl, getBotGateway), nil)
	req.Header.Set("Authorization", "Bot "+b.Cfg.Token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != 200 {
		return fmt.Errorf("invalid bot token")
	}
	bytes, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}
	var botGateway botGatewayResp
	if err := json.Unmarshal(bytes, &botGateway); err != nil {
		return err
	}
	slog.Debug("bot_gateway", "url", botGateway.URL)
	b.gatewayUrl = botGateway.URL
	return nil
}
