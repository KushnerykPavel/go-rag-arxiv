package app

type Config struct {
	Address     string `envconfig:"ADDRESS" default:":8080"`
	GRPCAddress string `envconfig:"GRPC_ADDRESS" default:":9090"`
	GroqAPIKey  string `envconfig:"GROQ_API_KEY" required:"true"`

	TelegramConfig
}

type TelegramConfig struct {
	Token  string `envconfig:"TELEGRAM_TOKEN" required:"true"`
	ChatID int64  `envconfig:"TELEGRAM_CHAT_ID" required:"true"`
}
