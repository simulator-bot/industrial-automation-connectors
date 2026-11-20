package main

import (
	"log/slog"
	"os"

	"github.com/simulator-bot/industrial-automation-connectors/internal/gateway"
	"github.com/simulator-bot/industrial-automation-connectors/internal/platform/config"
)

func main() {
	cfg := config.Load("GATEWAY", ":8081")
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := gateway.NewServer(cfg, logger).Run(); err != nil {
		logger.Error("gateway server stopped", "error", err)
		os.Exit(1)
	}
}
