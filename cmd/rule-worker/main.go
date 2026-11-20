package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

// The rule worker intentionally has no device-network access. CEL evaluation
// and bounded action-graph execution will be added behind this process boundary.
func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()
	logger.Info("rule worker started", "state", "idle")
	<-ctx.Done()
	logger.Info("rule worker stopping")
}
