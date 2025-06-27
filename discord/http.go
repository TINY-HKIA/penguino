package discord

const (
	httpApiBaseUrl string = "https://discord.com/api/v10"
	getBotGateway  string = "/gateway/bot"
)

type botGatewayResp struct {
	URL               string `json:"url"`
	SessionStartLimit struct {
		MaxConcurrency int `json:"max_concurrency"`
		Remaining      int `json:"remaining"`
		ResetAfter     int `json:"reset_after"`
		Total          int `json:"total"`
	} `json:"session_start_limit"`
	Shards int `json:"shards"`
}
