package config

import "os"

type Config struct {
	ServiceName string
	HTTPAddress string
}

func Load(prefix, defaultAddress string) Config {
	name := os.Getenv(prefix + "_SERVICE_NAME")
	if name == "" {
		name = prefix
	}
	address := os.Getenv(prefix + "_HTTP_ADDR")
	if address == "" {
		address = defaultAddress
	}
	return Config{ServiceName: name, HTTPAddress: address}
}
