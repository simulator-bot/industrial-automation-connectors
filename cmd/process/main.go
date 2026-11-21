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
	server, err := process.NewServer(cfg, logger, process.Options{
		CatalogPath: os.Getenv("PROCESS_CATALOG_PATH"),
		GatewayURL:  os.Getenv("PROCESS_GATEWAY_URL"),
	})
	if err != nil {
		logger.Error("process startup failed", "error", err)
		os.Exit(1)
	}
	if err := server.Run(); err != nil {
		logger.Error("process server stopped", "error", err)
		os.Exit(1)
	}
}
