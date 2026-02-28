package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/d6o/homeclip/internal/cleanup"
	"github.com/d6o/homeclip/internal/clipboard"
	"github.com/d6o/homeclip/internal/config"
	"github.com/d6o/homeclip/internal/filestore"
	"github.com/d6o/homeclip/internal/server"
)

func main() {
	cfg := config.NewConfig()

	if err := os.MkdirAll(cfg.DataDir, 0o755); err != nil {
		slog.Error("failed to create data directory", "error", err)
		os.Exit(1)
	}

	clipStore := clipboard.NewStore(cfg.DataDir)

	fileStore, err := filestore.NewStore(cfg.DataDir)
	if err != nil {
		slog.Error("failed to create file store", "error", err)
		os.Exit(1)
	}

	cleaner := cleanup.NewCleaner(10*time.Minute, 24*time.Hour, clipStore, fileStore)
	srv := server.NewServer(cfg.Port, clipStore, fileStore)

	sigCtx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	g, gCtx := errgroup.WithContext(sigCtx)

	g.Go(func() error {
		return cleaner.Run(gCtx)
	})

	g.Go(func() error {
		return srv.Run(gCtx)
	})

	slog.Info("homeclip starting", "port", cfg.Port, "dataDir", cfg.DataDir)

	if err := g.Wait(); err != nil && sigCtx.Err() == nil {
		slog.Error("homeclip exited with error", "error", err)
		os.Exit(1)
	}

	slog.Info("homeclip stopped")
}
