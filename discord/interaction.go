package discord

// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object-interaction-callback-type
type InteractionType int

var (
	Msg InteractionType = 4
)

type Message struct {
	// Message content
	Content string `json:"content"` //
	// 	Supports up to 10 embeds
	Embeds []*Embed `json:"embeds"`
}

// https://discord.com/developers/docs/resources/message#embed-object
type Embed struct {
	// 	title of embed
	Title string `json:"title,omitempty"`
	// description of embed
	Description string `json:"description,omitempty"`
	// url of embed
	Url string `json:"url,omitempty"`
	// color code of the embed
	Color int `json:"color,omitempty"`
}

type interactionCreate struct {
	ID    string `json:"id"`
	Type  int    `json:"type"`
	Token string `json:"token"`
	Data  struct {
		Type int    `json:"type"`
		Name string `json:"name"`
	} `json:"data"`
}

// https://discord.com/developers/docs/interactions/receiving-and-responding#interaction-response-object
type InteractionResponse struct {
	Type InteractionType      `json:"type"`
	Data *Message `json:"data,omitempty"`
}
