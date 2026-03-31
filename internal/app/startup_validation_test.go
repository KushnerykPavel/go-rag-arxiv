package app

import (
	"fmt"
	"strings"
	"testing"
)

func TestStartupValidation(t *testing.T) {
	t.Parallel()

	cfg := Config{}

	err := startupValidate(cfg)
	if err == nil {
		t.Fatal("startupValidate() error = nil, want non-nil")
	}

	if !strings.Contains(err.Error(), "startup config validation failed") {
		t.Fatalf("startupValidate() error = %q, missing startup context", err.Error())
	}

	if !strings.Contains(err.Error(), "missing required runtime config: GROQ_API_KEY") {
		t.Fatalf("startupValidate() error = %q, missing runtime key detail", err.Error())
	}
}

func startupValidate(cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return fmt.Errorf("startup config validation failed: %w", err)
	}

	return nil
}
