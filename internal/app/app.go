package app

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/klyakssa/aggregation-sub/internal/config"
	"github.com/klyakssa/aggregation-sub/internal/logger"
	"github.com/klyakssa/aggregation-sub/internal/repository/postgres"
	"github.com/klyakssa/aggregation-sub/internal/service"
	httptransport "github.com/klyakssa/aggregation-sub/internal/transport/http"
	"go.uber.org/zap"
)

func Run(cfg *config.Config) {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger := logger.NewLogger(cfg.Logging, cfg.App.Name)
	logger.Info("Config initialized:")
	logger.Info("Debug mode: " + fmt.Sprintf("%v", cfg.Debug))
	logger.Info("App name: " + cfg.App.Name)

	logger.Info("Initializing database connection...")
	repo, err := postgres.NewPostgresStorage(cfg, logger)
	if err != nil {
		logger.Error("Failed to initialize database connection", zap.Error(err))
		return
	}

	logger.Info("Starting application...")

	// сервис
	subsService := service.NewSubsService(repo)

	// handler
	subsHandler := httptransport.NewSubscriptionHandler(logger, subsService)

	// router
	router := httptransport.NewRouter(logger, cfg)
	router.RegisterRoutes(subsHandler)

	go func() {
		if err := router.Run(ctx); err != nil && errors.Is(err, context.Canceled) {
			logger.Error("HTTP server stopped unexpectedly", zap.Error(err))
			cancel()
		}
	}()

	logger.Info("Application started")

	<-ctx.Done()
	logger.Info("Shutting down application...")

	if err := repo.Close(); err != nil {
		logger.Error("Failed to close database connection", zap.Error(err))
	}

	logger.Info("Application terminated gracefully")
}
