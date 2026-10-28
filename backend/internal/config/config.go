package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	Port        string
	CORSOrigin  string
}

func Load() Config {
	cfg := Config{
		DatabaseURL: env("DATABASE_URL", "postgres://stockflow:stockflow@localhost:5433/stockflow?sslmode=disable"),
		Port:        env("PORT", "8081"),
		CORSOrigin:  env("CORS_ORIGIN", "http://localhost:5173"),
	}
	return cfg
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
