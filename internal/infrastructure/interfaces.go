package infrastructure

import (
	"context"
	"net/http"
	"time"
)

//go:generate go tool mockgen -source=interfaces.go -destination=mocks.go -package=infrastructure

type HTTPServer interface {
	Start() error
	Shutdown(ctx context.Context) error
	Port() string
	URL() string
}

type Router interface {
	ServeHTTP(w http.ResponseWriter, r *http.Request)
}

type CleanupService interface {
	Start(ctx context.Context)
	Stop()
}

type ConfigLoader interface {
	LoadConfig() *Config
}

type Config struct {
	Port                  string
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	IdleTimeout           time.Duration
	MaxContentSize        int64
	MaxFileSize           int64
	MaxFileNameLength     int
	ExpirationDuration    time.Duration
	CleanupInterval       time.Duration
	ExpirationGracePeriod time.Duration
	EnableFileUploads     bool
	EnableAutoCleanup     bool
}
