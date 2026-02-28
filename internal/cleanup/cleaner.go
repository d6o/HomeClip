package cleanup

import (
	"context"
	"log/slog"
	"time"
)

type cleanable interface {
	Cleanup(ctx context.Context, maxAge time.Duration) error
}

type Cleaner struct {
	targets  []cleanable
	interval time.Duration
	maxAge   time.Duration
}

func NewCleaner(interval, maxAge time.Duration, targets ...cleanable) *Cleaner {
	return &Cleaner{
		targets:  targets,
		interval: interval,
		maxAge:   maxAge,
	}
}

func (c *Cleaner) Run(ctx context.Context) error {
	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
			c.runCycle(ctx)
		}
	}
}

func (c *Cleaner) runCycle(ctx context.Context) {
	for _, t := range c.targets {
		if err := t.Cleanup(ctx, c.maxAge); err != nil {
			slog.Error("cleanup failed", "error", err)
		}
	}

	slog.Info("cleanup cycle completed")
}
