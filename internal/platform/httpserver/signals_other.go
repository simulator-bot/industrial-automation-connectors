//go:build !unix

package httpserver

import (
	"os"
	"os/signal"
)

func shutdownSignal() <-chan os.Signal {
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	return ch
}
