package feishu

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// WriteWait is the time allowed to write a message to the peer.
	WriteWait = 10 * time.Second

	// PongWait is the time allowed to read the next pong message from the peer.
	PongWait = 60 * time.Second

	// PingPeriod is the interval for sending ping messages.
	PingPeriod = (PongWait * 9) / 10

	// MaxMessageSize is the maximum size of a message.
	MaxMessageSize = 65536
)

// WebSocketClient manages WebSocket connection to Feishu.
type WebSocketClient struct {
	config     *Config
	conn       *websocket.Conn
	handler    *EventHandler
	verifier   *SignatureVerifier

	mu         sync.RWMutex
	connected  bool
	token      string
}

// NewWebSocketClient creates a new WebSocket client.
func NewWebSocketClient(cfg *Config, handler *EventHandler) *WebSocketClient {
	return &WebSocketClient{
		config:  cfg,
		handler: handler,
		verifier: NewSignatureVerifier(cfg.VerificationToken, cfg.WebhookSecret),
	}
}

// Connect establishes WebSocket connection.
func (ws *WebSocketClient) Connect(ctx context.Context) error {
	dialer := websocket.Dialer{
		HandshakeTimeout: 10 * time.Second,
	}

	conn, _, err := dialer.DialContext(ctx, ws.config.WebSocketURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}

	ws.mu.Lock()
	ws.conn = conn
	ws.connected = true
	ws.mu.Unlock()

	// Configure connection
	conn.SetReadLimit(MaxMessageSize)
	conn.SetReadDeadline(time.Now().Add(PongWait))
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(PongWait))
		return nil
	})

	// Start ping ticker
	go ws.pingLoop()

	// Start message handler
	go ws.messageLoop(ctx)

	return nil
}

// Disconnect closes the WebSocket connection.
func (ws *WebSocketClient) Disconnect() error {
	ws.mu.Lock()
	defer ws.mu.Unlock()

	if !ws.connected || ws.conn == nil {
		return nil
	}

	// Send close frame
	err := ws.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	if err != nil {
		return err
	}

	// Wait for close or timeout
	ws.conn.SetReadDeadline(time.Now().Add(WriteWait))
	time.Sleep(100 * time.Millisecond)

	// Close connection
	ws.conn.Close()

	ws.connected = false
	ws.conn = nil

	return nil
}

// IsConnected returns whether the client is connected.
func (ws *WebSocketClient) IsConnected() bool {
	ws.mu.RLock()
	defer ws.mu.RUnlock()
	return ws.connected
}

// messageLoop reads and processes messages from WebSocket.
func (ws *WebSocketClient) messageLoop(ctx context.Context) {
	defer ws.Disconnect()

	for {
		select {
		case <-ctx.Done():
			return
		default:
			if !ws.IsConnected() {
				return
			}

			ws.mu.RLock()
			conn := ws.conn
			ws.mu.RUnlock()

			if conn == nil {
				return
			}

			// Read message
			messageType, data, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					ws.logError("read error: %v", err)
				}
				return
			}

			if messageType == websocket.CloseMessage {
				ws.logInfo("received close message")
				return
			}

			// Process text messages
			if messageType == websocket.TextMessage {
				ws.handleMessage(ctx, data)
			}
		}
	}
}

// handleMessage processes an incoming WebSocket message.
func (ws *WebSocketClient) handleMessage(ctx context.Context, data []byte) {
	var wsMsg WebSocketMessage
	if err := json.Unmarshal(data, &wsMsg); err != nil {
		ws.logError("failed to decode message: %v", err)
		return
	}

	// Handle different event types
	switch wsMsg.Header.EventType {
	case "im":
		ws.handleIMMessage(ctx, &wsMsg)
	case "tenant":
		ws.handleTenantMessage(ctx, &wsMsg)
	case "workspace":
		ws.handleWorkspaceMessage(ctx, &wsMsg)
	case "bot":
		ws.handleBotMessage(ctx, &wsMsg)
	default:
		ws.logError("unknown event type: %s", wsMsg.Header.EventType)
	}
}

// handleIMMessage processes IM messages.
func (ws *WebSocketClient) handleIMMessage(ctx context.Context, wsMsg *WebSocketMessage) {
	var msgEvent Event
	if err := json.Unmarshal(wsMsg.Data, &msgEvent); err != nil {
		ws.logError("failed to decode IM message: %v", err)
		return
	}

	if msgEvent.EventType != "im.message.receive_v1" {
		return
	}

	// Convert to incoming message
	msg, err := ws.parseMessageEvent(&msgEvent)
	if err != nil {
		ws.logError("failed to parse message: %v", err)
		return
	}

	// Handle message
	ws.handler.HandleMessage(ctx, msg)
}

// handleTenantMessage processes tenant events.
func (ws *WebSocketClient) handleTenantMessage(ctx context.Context, wsMsg *WebSocketMessage) {
	// Handle tenant token refresh, etc.
	ws.logInfo("received tenant event: %s", wsMsg.Header.EventType)
}

// handleWorkspaceMessage processes workspace events.
func (ws *WebSocketClient) handleWorkspaceMessage(ctx context.Context, wsMsg *WebSocketMessage) {
	// Handle workspace events
	ws.logInfo("received workspace event: %s", wsMsg.Header.EventType)
}

// handleBotMessage processes bot events.
func (ws *WebSocketClient) handleBotMessage(ctx context.Context, wsMsg *WebSocketMessage) {
	// Handle bot events
	ws.logInfo("received bot event: %s", wsMsg.Header.EventType)
}

// parseMessageEvent parses a message event into IncomingMessage.
func (ws *WebSocketClient) parseMessageEvent(event *Event) (*IncomingMessage, error) {
	if event.Event.Message == nil {
		return nil, fmt.Errorf("no message in event")
	}

	msgEvent := event.Event.Message

	// Parse message content
	var content string
	var images []ImageAttachment
	var files []FileAttachment

	switch msgEvent.Message.MessageType {
	case "text":
		var textContent TextMessageContent
		if err := json.Unmarshal(msgEvent.Message.Content, &textContent); err != nil {
			return nil, fmt.Errorf("failed to parse text content: %w", err)
		}
		content = textContent.Text

	case "image":
		var imgContent ImageMessageContent
		if err := json.Unmarshal(msgEvent.Message.Content, &imgContent); err != nil {
			return nil, fmt.Errorf("failed to parse image content: %w", err)
		}
		images = append(images, ImageAttachment{
			ImageKey: imgContent.ImageKey,
		})

	case "file":
		var fileContent FileMessageContent
		if err := json.Unmarshal(msgEvent.Message.Content, &fileContent); err != nil {
			return nil, fmt.Errorf("failed to parse file content: %w", err)
		}
		files = append(files, FileAttachment{
			FileKey: fileContent.FileKey,
		})

	default:
		return nil, fmt.Errorf("unsupported message type: %s", msgEvent.Message.MessageType)
	}

	// Build incoming message
	msg := &IncomingMessage{
		ChatID:     msgEvent.ChatID,
		MessageID:   msgEvent.MessageID,
		RootID:      msgEvent.RootID,
		ParentID:    msgEvent.ParentID,
		UserID:      msgEvent.Sender.UserID,
		UserName:    msgEvent.Sender.UserID,
		TenantKey:   msgEvent.Sender.TenantKey,
		Content:     content,
		Images:      images,
		Files:       files,
		Timestamp:   time.Unix(msgEvent.CreateTime, 0),
	}

	return msg, nil
}

// pingLoop sends periodic pings to keep connection alive.
func (ws *WebSocketClient) pingLoop() {
	ticker := time.NewTicker(PingPeriod)
	defer ticker.Stop()

	for range ticker.C {
		if !ws.IsConnected() {
			return
		}

		ws.mu.RLock()
		conn := ws.conn
		ws.mu.RUnlock()

		if conn == nil {
			return
		}

		conn.SetWriteDeadline(time.Now().Add(WriteWait))
		if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
			ws.logError("ping error: %v", err)
			ws.Disconnect()
			return
		}
	}
}

// logError logs an error message.
func (ws *WebSocketClient) logError(format string, args ...any) {
	fmt.Printf("[WebSocket] ERROR: "+format+"\n", args...)
}

// logInfo logs an info message.
func (ws *WebSocketClient) logInfo(format string, args ...any) {
	fmt.Printf("[WebSocket] INFO: "+format+"\n", args...)
}