package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kelseyhightower/envconfig"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/KushnerykPavel/go-rag-arxiv/internal/app"
)

func main() {
	var cfg app.Config
	err := envconfig.Process("arxiv-rag-go", &cfg)
	if err != nil {
		log.Fatal(err.Error())
	}

	config := zap.NewProductionConfig()
	config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC3339)

	l, _ := config.Build()
	logger := l.Sugar().With(zap.String("service", "arxiv-rag-go"))
	defer func() {
		if err := logger.Sync(); err != nil {
			log.Printf("logger: sync: %s", err) // nolint:forbidigo // it's ok here
		}
	}()
	ctx, ctxCancel := context.WithCancel(context.Background())
	go handleSignals(ctxCancel, logger)

	if err := app.New(cfg, logger).Run(ctx); err != nil {
		logger.Fatal("application run failed", "error", err)
	}
}

func handleSignals(cancel context.CancelFunc, l *zap.SugaredLogger) {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGTERM, syscall.SIGINT)

	<-signals
	l.Info("got signal; canceling context")
	cancel()

	<-signals
	l.Warn("got second signal; force exiting")
	if err := l.Sync(); err != nil {
		log.Printf("logger: sync: %s", err) // nolint:forbidigo // it's ok here
	}

	os.Exit(1)
}
