package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/requel1/price-tracker/internal/config"
	"github.com/requel1/price-tracker/internal/repository"
	"github.com/requel1/price-tracker/internal/service"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg, err := config.Load()
	if err != nil {
		slog.Error("config", "err", err)
		os.Exit(1)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	pool, err := pgxpool.New(ctx, cfg.DBURL())
	if err != nil {
		slog.Error("db pool", "err", err)
		os.Exit(1)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		slog.Error("db ping", "err", err)
		os.Exit(1)
	}
	slog.Info("connected to database")

	productRepo := repository.NewProductRepo(pool)
	priceRepo := repository.NewPriceRepo(pool)

	svc := service.NewParserService(productRepo, priceRepo, 5)

	interval := 6 * time.Hour
	slog.Info("worker started", "interval", interval)
	svc.RunBySchedule(ctx, interval)

	slog.Info("worker stopped")
}
