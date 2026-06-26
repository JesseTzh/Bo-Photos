package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Runtime struct {
	Address         string
	DataDir         string
	DatabasePath    string
	FrontendDir     string
	CookieSecure    bool
	MaxUploadBytes  int64
	InitialPassword string
}

func Load() (Runtime, error) {
	dataDir := envOrDefault("BOPHOTOS_DATA_DIR", "/data")
	cookieSecure, err := boolEnv("BOPHOTOS_COOKIE_SECURE", false)
	if err != nil {
		return Runtime{}, err
	}
	maxUploadBytes, err := int64Env("BOPHOTOS_MAX_UPLOAD_BYTES", 2<<30)
	if err != nil {
		return Runtime{}, err
	}

	return Runtime{
		Address:         envOrDefault("BOPHOTOS_ADDR", ":8080"),
		DataDir:         dataDir,
		DatabasePath:    filepath.Join(dataDir, "app.db"),
		FrontendDir:     envOrDefault("BOPHOTOS_FRONTEND_DIR", "frontend/dist"),
		CookieSecure:    cookieSecure,
		MaxUploadBytes:  maxUploadBytes,
		InitialPassword: os.Getenv("BOPHOTOS_INITIAL_PASSWORD"),
	}, nil
}

func int64Env(key string, fallback int64) (int64, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil || parsed <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", key)
	}
	return parsed, nil
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func boolEnv(key string, fallback bool) (bool, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return false, fmt.Errorf("%s must be a boolean: %w", key, err)
	}
	return parsed, nil
}
