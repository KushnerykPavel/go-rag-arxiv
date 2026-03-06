package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// Client sends notifications via the Telegram Bot API.
type Client struct {
	cfg   Config
	token string
	log   *zap.SugaredLogger
	http  *http.Client
}

// NewClient creates a Telegram bot client.
// token is the bot token issued by @BotFather.
func NewClient(token string, log *zap.SugaredLogger, opts ...Option) *Client {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.timeout}
	}

	return &Client{
		cfg:   cfg,
		token: token,
		log:   log.With("client", "telegram"),
		http:  httpClient,
	}
}

// SendMessage sends a plain-text message to the given chat.
func (c *Client) SendMessage(ctx context.Context, chatID int64, text string) error {
	return c.send(ctx, chatID, text, "")
}

// SendHTML sends an HTML-formatted message to the given chat.
// Telegram supports a limited HTML subset: <b>, <i>, <u>, <s>, <code>, <pre>, <a>.
func (c *Client) SendHTML(ctx context.Context, chatID int64, html string) error {
	return c.send(ctx, chatID, html, ParseModeHTML)
}

// SendMarkdown sends a MarkdownV2-formatted message to the given chat.
func (c *Client) SendMarkdown(ctx context.Context, chatID int64, md string) error {
	return c.send(ctx, chatID, md, ParseModeMarkdownV2)
}

// --- internal ---

type sendMessageRequest struct {
	ChatID    int64  `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode,omitempty"`
}

type apiResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

func (c *Client) send(ctx context.Context, chatID int64, text, parseMode string) error {
	payload := sendMessageRequest{
		ChatID:    chatID,
		Text:      text,
		ParseMode: parseMode,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshaling request: %w", err)
	}

	endpoint := fmt.Sprintf("%s/bot%s/sendMessage", c.cfg.baseURL, c.token)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("decoding response: %w", err)
	}

	if !apiResp.OK {
		return fmt.Errorf("telegram API error: %s", apiResp.Description)
	}

	c.log.Infow("notification sent", "chat_id", chatID)
	return nil
}
