package main

import (
	"log/slog"
	"os"

	"github.com/simulator-bot/industrial-automation-connectors/internal/platform/config"
	"github.com/simulator-bot/industrial-automation-connectors/internal/process"
)

func main() {
	cfg := config.Load("PROCESS", ":8080")
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	if err := process.NewServer(cfg, logger).Run(); err != nil {
		logger.Error("process server stopped", "error", err)
		os.Exit(1)
	}
}
