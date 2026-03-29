package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/dvnpv/subscriptions-aggregator/docs"
	"github.com/dvnpv/subscriptions-aggregator/internal/config"
	"github.com/dvnpv/subscriptions-aggregator/internal/handler"
	"github.com/dvnpv/subscriptions-aggregator/internal/middleware"
	"github.com/dvnpv/subscriptions-aggregator/internal/repository"
	"github.com/dvnpv/subscriptions-aggregator/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	httpSwagger "github.com/swaggo/http-swagger"
)

// @title Subscription Service API
// @version 1.0
// @description REST-сервис для агрегации данных об онлайн-подписках пользователей
// @BasePath /
func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.LogLevel,
	}))

	ctx := context.Background()

	dbpool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		logger.Error("failed to create db pool", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer dbpool.Close()

	if err = dbpool.Ping(ctx); err != nil {
		logger.Error("failed to ping db", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("database connected")

	repo := repository.NewSubscriptionRepository(dbpool)
	svc := service.NewSubscriptionService(repo)
	h := handler.NewSubscriptionHandler(svc, logger)
	healthHandler := handler.NewHealthHandler()

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger(logger))

	r.Get("/health", healthHandler.Health)
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/subscriptions", func(r chi.Router) {
		r.Post("/", h.Create)
		r.Get("/", h.List)
		r.Get("/total", h.GetTotal)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})

	srv := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           r,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		logger.Info("server started", slog.String("addr", cfg.HTTPAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	logger.Info("shutting down server")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("server stopped gracefully")
}
