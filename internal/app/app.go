package app

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-co-op/gocron/v2"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc"

	"github.com/KushnerykPavel/go-rag-arxiv/internal/client/arxiv"
	"github.com/KushnerykPavel/go-rag-arxiv/internal/client/groq"
	"github.com/KushnerykPavel/go-rag-arxiv/internal/client/telegram"
	"github.com/KushnerykPavel/go-rag-arxiv/internal/cron"
	arxivv1 "github.com/KushnerykPavel/go-rag-arxiv/internal/gen/arxiv/v1"
	"github.com/KushnerykPavel/go-rag-arxiv/internal/rag"
	grpcserver "github.com/KushnerykPavel/go-rag-arxiv/internal/server/grpc"
	"github.com/KushnerykPavel/go-rag-arxiv/internal/wrappers"
)

const shutdownTimeout = 5 * time.Second

type App struct {
	cfg Config
	l   *zap.SugaredLogger
}

func New(cfg Config, l *zap.SugaredLogger) *App {
	return &App{cfg: cfg, l: l}
}

func (a *App) Run(ctx context.Context) error {
	errGrp, _ := errgroup.WithContext(ctx)

	scheduler, err := gocron.NewScheduler()
	if err != nil {
		return fmt.Errorf("creating scheduler: %w", err)
	}

	arxivLimiter, err := wrappers.NewRateLimiter(1)
	if err != nil {
		return fmt.Errorf("creating arxiv rate limiter: %w", err)
	}

	arxivClient := arxiv.NewClient(a.l)
	groqClient := groq.NewClient(a.cfg.GroqAPIKey, a.l)
	askService := rag.NewAskService(arxivClient, groqClient)
	telegramClient := telegram.NewClient(a.cfg.TelegramConfig.Token, a.l)
	arxivFetcher := cron.NewArxivFetcher(arxivClient, telegramClient, a.cfg.TelegramConfig.ChatID, a.l, arxivLimiter)

	_, err = scheduler.NewJob(
		gocron.CronJob("0 5 * * *", false),
		gocron.NewTask(arxivFetcher.FetchPapers, ctx),
	)
	if err != nil {
		return fmt.Errorf("registering arxiv fetch job: %w", err)
	}

	scheduler.Start()
	a.l.Infow("scheduler started", "schedule", "daily at 05:00")
	_ = telegramClient.SendMarkdown(ctx, a.cfg.TelegramConfig.ChatID, "🚀 application started")

	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	srv := &http.Server{
		Addr:    a.cfg.Address,
		Handler: r,
	}

	errGrp.Go(func() error {
		a.l.Infow("http server started", "address", a.cfg.Address)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server: %w", err)
		}
		return nil
	})

	// gRPC server
	grpcSrv := grpc.NewServer()
	arxivv1.RegisterArxivServiceServer(grpcSrv, grpcserver.NewArxivHandler(arxivClient, askService, a.l))

	errGrp.Go(func() error {
		lis, err := net.Listen("tcp", a.cfg.GRPCAddress)
		if err != nil {
			return fmt.Errorf("grpc listen: %w", err)
		}
		a.l.Infow("grpc server started", "address", a.cfg.GRPCAddress)
		if err := grpcSrv.Serve(lis); err != nil {
			return fmt.Errorf("grpc server: %w", err)
		}
		return nil
	})

	<-ctx.Done()
	a.l.Info("shutting down")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := scheduler.Shutdown(); err != nil {
		return fmt.Errorf("shutting down scheduler: %w", err)
	}
	if err := srv.Shutdown(shutdownCtx); err != nil {
		a.l.Errorw("http server shutdown error", "error", err)
	}
	grpcSrv.GracefulStop()

	_ = telegramClient.SendMarkdown(ctx, a.cfg.TelegramConfig.ChatID, "🚀 application stopped")
	return errGrp.Wait()
}
