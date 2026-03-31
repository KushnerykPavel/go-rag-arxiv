package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"go.uber.org/zap"
)

const chatCompletionsEndpoint = "https://api.groq.com/openai/v1/chat/completions"

type Client struct {
	cfg    Config
	apiKey string
	log    *zap.SugaredLogger
	http   *http.Client
}

func NewClient(apiKey string, log *zap.SugaredLogger, opts ...Option) *Client {
	cfg := defaultConfig()
	for _, opt := range opts {
		opt(&cfg)
	}

	httpClient := cfg.httpClient
	if httpClient == nil {
		httpClient = &http.Client{Timeout: cfg.timeout}
	}
	if cfg.chatCompletionsURL == "" {
		cfg.chatCompletionsURL = chatCompletionsEndpoint
	}

	return &Client{
		cfg:    cfg,
		apiKey: apiKey,
		log:    log.With("client", "groq"),
		http:   httpClient,
	}
}

func (c *Client) GenerateAnswer(ctx context.Context, query string, contextBlock string) (string, error) {
	payload := chatCompletionRequest{
		Model: c.cfg.model,
		Messages: []chatCompletionMessage{
			{
				Role: "system",
				Content: "You answer only from provided arXiv context. " +
					"If context is insufficient, state uncertainty briefly.",
			},
			{
				Role:    "user",
				Content: fmt.Sprintf("Question:\n%s\n\nContext:\n%s", query, contextBlock),
			},
		},
		MaxTokens:   c.cfg.maxTokens,
		Temperature: c.cfg.temperature,
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshaling request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.cfg.chatCompletionsURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("executing request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		errorBody, _ := io.ReadAll(resp.Body)
		message := extractGroqError(errorBody)
		return "", APIError{
			statusCode: resp.StatusCode,
			message:    message,
		}
	}

	var completion chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&completion); err != nil {
		return "", fmt.Errorf("decoding response: %w", err)
	}

	if len(completion.Choices) == 0 {
		return "", fmt.Errorf("empty choices in completion response")
	}

	answer := strings.TrimSpace(completion.Choices[0].Message.Content)
	if answer == "" {
		return "", fmt.Errorf("empty answer in completion response")
	}

	c.log.Debugw("generated ask answer", "model", c.cfg.model)
	return answer, nil
}

type APIError struct {
	statusCode int
	message    string
}

func (e APIError) Error() string {
	return fmt.Sprintf("groq API error: status=%d message=%s", e.statusCode, e.message)
}

func (e APIError) StatusCode() int {
	return e.statusCode
}

type chatCompletionRequest struct {
	Model       string                  `json:"model"`
	Messages    []chatCompletionMessage `json:"messages"`
	MaxTokens   int                     `json:"max_tokens,omitempty"`
	Temperature float32                 `json:"temperature,omitempty"`
}

type chatCompletionMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message chatCompletionMessage `json:"message"`
	} `json:"choices"`
}

type chatCompletionErrorResponse struct {
	Error struct {
		Message string `json:"message"`
	} `json:"error"`
}

func extractGroqError(body []byte) string {
	if len(body) == 0 {
		return "empty error response"
	}

	var apiErr chatCompletionErrorResponse
	if err := json.Unmarshal(body, &apiErr); err == nil && apiErr.Error.Message != "" {
		return apiErr.Error.Message
	}

	return strings.TrimSpace(string(body))
}
