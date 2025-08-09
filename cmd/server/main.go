package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/d6o/homeclip/internal/embed"
	"github.com/d6o/homeclip/internal/infrastructure/container"
)

func main() {
	container := container.NewContainer(embed.StaticFiles)

	// Start cleanup service if enabled
	ctx := context.Background()
	if container.Config.EnableAutoCleanup {
		container.CleanupService.Start(ctx)
	}

	// Setup graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		// Stop cleanup service if running
		if container.Config.EnableAutoCleanup {
			container.CleanupService.Stop()
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := container.Server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	if err := container.StartServer(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}