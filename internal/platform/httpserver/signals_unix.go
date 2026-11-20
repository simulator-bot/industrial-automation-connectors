//go:build unix

package httpserver

import (
	"os"
	"os/signal"
	"syscall"
)

func shutdownSignal() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM)
	return ch
}
