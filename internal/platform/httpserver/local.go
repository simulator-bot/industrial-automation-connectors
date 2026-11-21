package httpserver

import (
	"net"
	"net/http"
	"strings"
)

func IsLoopbackAddress(address string) bool {
	host, _, err := net.SplitHostPort(address)
	return err == nil && net.ParseIP(host).IsLoopback()
}

func LocalOnly(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			host = strings.Trim(r.Host, "[]")
		}
		if !strings.EqualFold(host, "localhost") && !net.ParseIP(host).IsLoopback() {
			http.Error(w, "local API requires a loopback Host", http.StatusForbidden)
			return
		}
		handler(w, r)
	}
}
