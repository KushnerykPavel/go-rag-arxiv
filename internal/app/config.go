package app

import "fmt"

type Config struct {
	Address     string `envconfig:"ADDRESS" default:":8080"`
	GRPCAddress string `envconfig:"GRPC_ADDRESS" default:":9090"`
	GroqAPIKey  string `envconfig:"GROQ_API_KEY"`

	TelegramConfig
}

type TelegramConfig struct {
	Token  string `envconfig:"TELEGRAM_TOKEN"`
	ChatID int64  `envconfig:"TELEGRAM_CHAT_ID"`
}

func (c Config) Validate() error {
	if c.GroqAPIKey == "" {
		return fmt.Errorf("missing required runtime config: GROQ_API_KEY")
	}

	if c.TelegramConfig.Token == "" {
		return fmt.Errorf("missing required runtime config: TELEGRAM_TOKEN")
	}

	if c.TelegramConfig.ChatID == 0 {
		return fmt.Errorf("missing required runtime config: TELEGRAM_CHAT_ID")
	}

	return nil
}
