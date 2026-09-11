package main

import "os"

type serveConfig struct {
	IndexPath string
	SharedDir string
	Addr      string
}

func envOrDefault(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}

func defaultServeConfig() serveConfig {
	return serveConfig{
		IndexPath: envOrDefault("GODROP_INDEX", ""),
		SharedDir: envOrDefault("GODROP_DIR", ""),
		Addr:      envOrDefault("GODROP_ADDR", "127.0.0.1:8080"),
	}
}
