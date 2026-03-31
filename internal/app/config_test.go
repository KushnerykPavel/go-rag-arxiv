package app

import "testing"

func TestConfigValidation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     Config
		wantErr string
	}{
		{
			name: "passes when groq and telegram values are set",
			cfg: Config{
				GroqAPIKey: "groq-key",
				TelegramConfig: TelegramConfig{
					Token:  "token",
					ChatID: 12345,
				},
			},
		},
		{
			name: "fails when groq key is missing",
			cfg: Config{
				TelegramConfig: TelegramConfig{
					Token:  "token",
					ChatID: 12345,
				},
			},
			wantErr: "missing required runtime config: GROQ_API_KEY",
		},
		{
			name: "fails when telegram token is missing",
			cfg: Config{
				GroqAPIKey: "groq-key",
				TelegramConfig: TelegramConfig{
					ChatID: 12345,
				},
			},
			wantErr: "missing required runtime config: TELEGRAM_TOKEN",
		},
		{
			name: "fails when telegram chat id is missing",
			cfg: Config{
				GroqAPIKey: "groq-key",
				TelegramConfig: TelegramConfig{
					Token: "token",
				},
			},
			wantErr: "missing required runtime config: TELEGRAM_CHAT_ID",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := tt.cfg.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Fatalf("Validate() error = %v, want nil", err)
				}
				return
			}

			if err == nil {
				t.Fatalf("Validate() error = nil, want %q", tt.wantErr)
			}

			if err.Error() != tt.wantErr {
				t.Fatalf("Validate() error = %q, want %q", err.Error(), tt.wantErr)
			}
		})
	}
}
