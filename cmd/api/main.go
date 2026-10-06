package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/requel1/price-tracker/internal/config"
	"github.com/requel1/price-tracker/internal/handler"
	"github.com/requel1/price-tracker/internal/repository"
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

	ctx := context.Background()
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

	productHandler := handler.NewProductHandler(productRepo, priceRepo)
	webHandler := handler.NewWebHandler("web")

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/products", productHandler.List)
	mux.HandleFunc("POST /api/products", productHandler.Create)
	mux.HandleFunc("GET /api/products/{id}", productHandler.GetByID)
	mux.HandleFunc("DELETE /api/products/{id}", productHandler.Delete)
	mux.HandleFunc("GET /api/products/{id}/prices", productHandler.GetPrices)

	mux.HandleFunc("GET /", webHandler.Index)

	srv := &http.Server{
		Addr:         ":" + cfg.APIPort,
		Handler:      mux,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	go func() {
		slog.Info("server started", "port", cfg.APIPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server", "err", err)
			os.Exit(1)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	slog.Info("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("shutdown", "err", err)
	}
	slog.Info("stopped")
}
