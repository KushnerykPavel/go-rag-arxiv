package telegram

import (
	"net/http"
	"time"
)

const (
	// ParseModeHTML enables Telegram's HTML subset for message formatting.
	ParseModeHTML = "HTML"
	// ParseModeMarkdownV2 enables Telegram's MarkdownV2 formatting.
	ParseModeMarkdownV2 = "MarkdownV2"

	defaultBaseURL = "https://api.telegram.org"
	defaultTimeout = 10 * time.Second
)

// Config holds configuration for the Telegram bot client.
type Config struct {
	baseURL    string
	timeout    time.Duration
	httpClient *http.Client
}

// Option configures the Telegram client.
type Option func(*Config)

func WithBaseURL(url string) Option {
	return func(c *Config) { c.baseURL = url }
}

func WithTimeout(d time.Duration) Option {
	return func(c *Config) { c.timeout = d }
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) { c.httpClient = client }
}

func defaultConfig() Config {
	return Config{
		baseURL: defaultBaseURL,
		timeout: defaultTimeout,
	}
}
