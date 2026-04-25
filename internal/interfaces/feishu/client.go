package feishu

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/hashicorp/go-retryablehttp"
)

const (
	// BaseURL is the Feishu/Lark API base URL.
	BaseURL = "https://open.feishu.cn/open-apis"

	// AuthAPI is the authentication API endpoint.
	AuthAPI = BaseURL + "/auth/v3/tenant_access_token/internal"

	// SendMessageAPI is the send message API endpoint.
	SendMessageAPI = BaseURL + "/im/v1/messages"

	// UploadImageAPI is the upload image API endpoint.
	UploadImageAPI = BaseURL + "/im/v1/images"

	// UploadFileAPI is the upload file API endpoint.
	UploadFileAPI = BaseURL + "/drive/v1/files/upload_all"

	// GetUserInfoAPI is the get user info API endpoint.
	GetUserInfoAPI = BaseURL + "/contact/v3/users"
)

// Client is a Feishu/Lark API client.
type Client struct {
	httpClient *retryablehttp.Client
	config     *Config

	token      string
	tokenExp   time.Time
}

// NewClient creates a new Feishu API client.
func NewClient(cfg *Config) *Client {
	return &Client{
		httpClient: retryablehttp.NewClient(),
		config:     cfg,
	}
}

// GetTenantAccessToken retrieves a tenant access token.
func (c *Client) GetTenantAccessToken(ctx context.Context) (string, error) {
	// Check if token is still valid
	if c.token != "" && time.Now().Before(c.tokenExp.Add(-5*time.Minute)) {
		return c.token, nil
	}

	reqBody := TenantAccessTokenRequest{
		AppID:     c.config.AppID,
		AppSecret: c.config.AppSecret,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := c.newRequest(ctx, "POST", AuthAPI, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result TenantAccessTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("API error: %s", result.Message)
	}

	c.token = result.Data.TenantAccessToken
	c.tokenExp = time.Now().Add(time.Duration(result.Data.ExpireIn) * time.Second)

	return c.token, nil
}

// SendMessage sends a message to a Feishu chat.
func (c *Client) SendMessage(ctx context.Context, req SendMessageRequest) (*SendMessageResponse, error) {
	token, err := c.GetTenantAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := c.newRequest(ctx, "POST", SendMessageAPI, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result SendMessageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return &result, nil
}

// SendText sends a text message.
func (c *Client) SendText(ctx context.Context, chatID, text string) error {
	content := TextMessageContent{Text: text}
	contentJSON, _ := json.Marshal(content)

	req := SendMessageRequest{
		ReceiveIDType: "chat_id",
		ReceiveID:     chatID,
		MsgType:       "text",
		Content:       string(contentJSON),
	}

	_, err := c.SendMessage(ctx, req)
	return err
}

// SendMarkdown sends a markdown-formatted message.
func (c *Client) SendMarkdown(ctx context.Context, chatID, markdown string) error {
	content := map[string]string{
		"text": markdown,
	}
	contentJSON, _ := json.Marshal(content)

	req := SendMessageRequest{
		ReceiveIDType: "chat_id",
		ReceiveID:     chatID,
		MsgType:       "post",
		Content:       fmt.Sprintf(`{"post":{"zh_cn":[[%s]]}}`, contentJSON),
	}

	_, err := c.SendMessage(ctx, req)
	return err
}

// SendCard sends an interactive card message.
func (c *Client) SendCard(ctx context.Context, chatID string, card *Card) error {
	contentJSON, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("failed to marshal card: %w", err)
	}

	req := SendMessageRequest{
		ReceiveIDType: "chat_id",
		ReceiveID:     chatID,
		MsgType:       "interactive",
		Content:       string(contentJSON),
	}

	_, err = c.SendMessage(ctx, req)
	return err
}

// SendImage sends an image message.
func (c *Client) SendImage(ctx context.Context, chatID, imageKey string) error {
	content := ImageMessageContent{ImageKey: imageKey}
	contentJSON, _ := json.Marshal(content)

	req := SendMessageRequest{
		ReceiveIDType: "chat_id",
		ReceiveID:     chatID,
		MsgType:       "image",
		Content:       string(contentJSON),
	}

	_, err := c.SendMessage(ctx, req)
	return err
}

// UploadImage uploads an image and returns the image key.
func (c *Client) UploadImage(ctx context.Context, data []byte) (string, error) {
	token, err := c.GetTenantAccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	reqBody := UploadImageRequest{
		ImageType: "message",
		Image:     data,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := c.newRequest(ctx, "POST", UploadImageAPI, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result UploadImageResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("API error: %s", result.Message)
	}

	return result.Data.ImageKey, nil
}

// UploadFile uploads a file and returns the file key.
func (c *Client) UploadFile(ctx context.Context, data []byte, filename string) (string, error) {
	token, err := c.GetTenantAccessToken(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get access token: %w", err)
	}

	reqBody := UploadFileRequest{
		FileType: "file",
		File:     data,
		FileName: filename,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := c.newRequest(ctx, "POST", UploadFileAPI, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result UploadFileResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Code != 0 {
		return "", fmt.Errorf("API error: %s", result.Message)
	}

	return result.Data.FileKey, nil
}

// GetUserInfo retrieves user information.
func (c *Client) GetUserInfo(ctx context.Context, userID string) (map[string]any, error) {
	token, err := c.GetTenantAccessToken(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get access token: %w", err)
	}

	url := fmt.Sprintf("%s/%s", GetUserInfoAPI, userID)
	httpReq, err := c.newRequest(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	var result struct {
		Code    int             `json:"code"`
		Message string          `json:"msg"`
		Data    map[string]any  `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if result.Code != 0 {
		return nil, fmt.Errorf("API error: %s", result.Message)
	}

	return result.Data, nil
}

// newRequest creates a new HTTP request with common headers.
func (c *Client) newRequest(ctx context.Context, method, url string, body io.Reader) (*retryablehttp.Request, error) {
	req, err := retryablehttp.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "claude-code-go/1.0")

	return req, nil
}

// SetHTTPClient sets a custom HTTP client.
func (c *Client) SetHTTPClient(client *retryablehttp.Client) {
	c.httpClient = client
}