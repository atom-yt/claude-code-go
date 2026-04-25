package feishu

import (
	"encoding/json"
	"time"
)

// Event represents a Feishu webhook/event message.
type Event struct {
	Token     string      `json:"token"`
	EventType string      `json:"type"`    // "url_verification" or "event_callback"
	Challenge string      `json:"challenge,omitempty"` // For URL verification
	Timestamp int64       `json:"ts"`
	Event     InnerEvent  `json:"event,omitempty"`
}

// InnerEvent is the inner event payload.
type InnerEvent struct {
	Type      string        `json:"type"` // "message" or "im.message.receive_v1"
	Timestamp int64         `json:"ts"`
	Message   *MessageEvent `json:"message,omitempty"`
}

// MessageEvent represents a message event from Feishu.
type MessageEvent struct {
	ChatID     string         `json:"chat_id"`
	MessageID  string         `json:"message_id"`
	RootID     string         `json:"root_id,omitempty"`
	ParentID   string         `json:"parent_id,omitempty"`
	CreateTime int64          `json:"create_time"`
	Sender     SenderInfo     `json:"sender"`
	Message    MessageContent `json:"message"`
	Mentions   []Mention      `json:"mentions,omitempty"`
}

// SenderInfo contains information about the message sender.
type SenderInfo struct {
	UserID     string `json:"user_id"`
	SenderType string `json:"sender_type"`
	TenantKey  string `json:"tenant_key"`
}

// MessageContent contains the message content.
type MessageContent struct {
	MessageType string          `json:"message_type"`
	Content     json.RawMessage `json:"content"`
}

// TextMessageContent is the content of a text message.
type TextMessageContent struct {
	Text string `json:"text"`
}

// ImageMessageContent is the content of an image message.
type ImageMessageContent struct {
	ImageKey string `json:"image_key"`
}

// FileMessageContent is the content of a file message.
type FileMessageContent struct {
	FileKey string `json:"file_key"`
}

// Mention represents a @mention in a message.
type Mention struct {
	ID      string `json:"id"`
	IDType  string `json:"id_type"`
	Name    string `json:"name"`
	TenantKey string `json:"tenant_key"`
}

// IncomingMessage represents a normalized incoming message.
type IncomingMessage struct {
	ChatID     string
	MessageID  string
	RootID     string
	ParentID   string
	UserID     string
	UserName   string
	TenantKey  string
	Content    string
	Images     []ImageAttachment
	Files      []FileAttachment
	CardAction *CardAction
	Timestamp  time.Time
}

// ImageAttachment represents an attached image.
type ImageAttachment struct {
	ImageKey string
	URL      string
	Width    int
	Height   int
	AltText  string
}

// FileAttachment represents an attached file.
type FileAttachment struct {
	FileKey string
	URL     string
	Name    string
	Size    int64
	Type    string
}

// CardAction represents an action triggered from an interactive card.
type CardAction struct {
	ActionID  string
	Action    string
	Value     map[string]any
	Timestamp time.Time
}

// SendMessageRequest represents a request to send a message via Feishu API.
type SendMessageRequest struct {
	ReceiveID   string      `json:"receive_id"`
	ReceiveIDType string    `json:"receive_id_type"` // "open_id", "user_id", "email", "union_id", "chat_id"
	MsgType     string      `json:"msg_type"`        // "text", "post", "image", "file", "interactive"
	Content     string      `json:"content"`         // JSON string
	UUID        string      `json:"uuid"`           // Optional, for idempotency
}

// SendMessageResponse represents a response from the send message API.
type SendMessageResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    struct {
		MessageID string `json:"msg_id"`
	} `json:"data"`
}

// TenantAccessTokenRequest represents a request to get tenant access token.
type TenantAccessTokenRequest struct {
	AppID     string `json:"app_id"`
	AppSecret string `json:"app_secret"`
}

// TenantAccessTokenResponse represents a response from the tenant access token API.
type TenantAccessTokenResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    struct {
		TenantAccessToken string    `json:"tenant_access_token"`
		Expire           int64     `json:"expire"`
		ExpireIn         int       `json:"expire_in"`
	} `json:"data"`
}

// WebSocketMessage represents a message received via WebSocket.
type WebSocketMessage struct {
	Header WebSocketHeader `json:"header"`
	Data   json.RawMessage `json:"data"`
}

// WebSocketHeader contains the WebSocket message header.
type WebSocketHeader struct {
	EventID    string `json:"event_id"`
	EventType  string `json:"event_type"` // "tenant", "workspace", "bot", "im"
	TenantKey  string `json:"tenant_key"`
	AppID      string `json:"app_id"`
	CreateTime int64  `json:"create_time"`
}

// CardElement represents a card element for interactive messages.
type CardElement struct {
	Tag       string                 `json:"tag"`      // "div", "hr", "img", "action", "column_set", etc.
	Text      *CardText              `json:"text,omitempty"`
	Actions   []*CardActionElement   `json:"actions,omitempty"`
	Columns   []*CardColumn          `json:"columns,omitempty"`
	Extra     map[string]any         `json:"extra,omitempty"`
}

// CardText represents text in a card.
type CardText struct {
	Tag      string `json:"tag"`    // "lark_md", "plain_text"
	Content  string `json:"content"`
	Lines    int    `json:"lines,omitempty"`
}

// CardActionElement represents an action button in a card.
type CardActionElement struct {
	Tag      string                 `json:"tag"`      // "button"
	Text     CardText               `json:"text"`
	Type     string                 `json:"type"`     // "primary", "default", "danger"
	URL      string                 `json:"url,omitempty"`
	Value    map[string]any         `json:"value,omitempty"`
}

// CardColumn represents a column in a multi-column layout.
type CardColumn struct {
	Weight   int              `json:"weight"` // 1-10
	Elements []*CardElement   `json:"elements"`
}

// Card represents an interactive card.
type Card struct {
	Header  *CardHeader        `json:"header,omitempty"`
	Elements []*CardElement    `json:"elements"`
}

// CardHeader represents the header of a card.
type CardHeader struct {
	Title    *CardText `json:"title"`
	Subtitle *CardText `json:"subtitle,omitempty"`
	Template string    `json:"template"` // "blue", "wathet", "turquoise", "green", "yellow", "orange", "red", "carmine", "violet", "purple", "indigo", "grey"
}

// UploadImageRequest represents a request to upload an image.
type UploadImageRequest struct {
	ImageType string `json:"image_type"` // "message", "avatar", "sticker"
	Image     []byte `json:"image"`     // Base64 encoded
}

// UploadImageResponse represents a response from the upload image API.
type UploadImageResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    struct {
		ImageKey string `json:"image_key"`
	} `json:"data"`
}

// UploadFileRequest represents a request to upload a file.
type UploadFileRequest struct {
	FileType string `json:"file_type"` // "file", "stream"
	File     []byte `json:"file"`      // Base64 encoded
	FileName string `json:"file_name"`
}

// UploadFileResponse represents a response from the upload file API.
type UploadFileResponse struct {
	Code    int    `json:"code"`
	Message string `json:"msg"`
	Data    struct {
		FileKey string `json:"file_key"`
	} `json:"data"`
}

// MediaType represents the type of media file.
type MediaType string

const (
	MediaTypeImage MediaType = "image"
	MediaTypeFile  MediaType = "file"
)