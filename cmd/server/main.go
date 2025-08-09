package main

import (
	"context"
	"errors"
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
	server := container.NewContainer(embed.StaticFiles)

	ctx := context.Background()
	if server.Config.EnableAutoCleanup {
		server.CleanupService.Start(ctx)
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		if server.Config.EnableAutoCleanup {
			server.CleanupService.Stop()
		}

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Server.Shutdown(shutdownCtx); err != nil {
			log.Printf("Server shutdown error: %v", err)
		}
	}()

	if err := server.StartServer(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal(err)
	}
}
