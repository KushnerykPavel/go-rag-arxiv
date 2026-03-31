package groq

import (
	"net/http"
	"time"
)

const (
	defaultChatCompletionsURL = "https://api.groq.com/openai/v1/chat/completions"
	defaultModel              = "llama-3.3-70b-versatile"
	defaultTimeout            = 30 * time.Second
	defaultMaxTokens          = 700
	defaultTemperature        = 0.2
)

type Config struct {
	chatCompletionsURL string
	model              string
	timeout            time.Duration
	maxTokens          int
	temperature        float32
	httpClient         *http.Client
}

type Option func(*Config)

func WithChatCompletionsURL(url string) Option {
	return func(c *Config) { c.chatCompletionsURL = url }
}

func WithModel(model string) Option {
	return func(c *Config) { c.model = model }
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) { c.timeout = timeout }
}

func WithMaxTokens(maxTokens int) Option {
	return func(c *Config) { c.maxTokens = maxTokens }
}

func WithTemperature(temperature float32) Option {
	return func(c *Config) { c.temperature = temperature }
}

func WithHTTPClient(client *http.Client) Option {
	return func(c *Config) { c.httpClient = client }
}

func defaultConfig() Config {
	return Config{
		chatCompletionsURL: defaultChatCompletionsURL,
		model:              defaultModel,
		timeout:            defaultTimeout,
		maxTokens:          defaultMaxTokens,
		temperature:        defaultTemperature,
	}
}
