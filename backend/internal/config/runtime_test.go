package config

import (
	"path/filepath"
	"testing"
)

func TestLoadUsesProductionDefaults(t *testing.T) {
	t.Setenv("BOPHOTOS_ADDR", "")
	t.Setenv("BOPHOTOS_DATA_DIR", "")
	t.Setenv("BOPHOTOS_FRONTEND_DIR", "")
	t.Setenv("BOPHOTOS_COOKIE_SECURE", "")
	t.Setenv("BOPHOTOS_MAX_UPLOAD_BYTES", "")
	t.Setenv("BOPHOTOS_INITIAL_PASSWORD", "")

	runtime, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if runtime.Address != ":8080" {
		t.Errorf("Address = %q, want :8080", runtime.Address)
	}
	if runtime.DataDir != "/data" {
		t.Errorf("DataDir = %q, want /data", runtime.DataDir)
	}
	if runtime.DatabasePath != filepath.Join("/data", "app.db") {
		t.Errorf("DatabasePath = %q, want /data/app.db", runtime.DatabasePath)
	}
	if runtime.FrontendDir != "frontend/dist" {
		t.Errorf("FrontendDir = %q, want frontend/dist", runtime.FrontendDir)
	}
	if runtime.CookieSecure {
		t.Error("CookieSecure = true, want false")
	}
	if runtime.MaxUploadBytes != 2<<30 {
		t.Errorf("MaxUploadBytes = %d, want %d", runtime.MaxUploadBytes, int64(2<<30))
	}
	if runtime.InitialPassword != "" {
		t.Errorf("InitialPassword = %q, want empty", runtime.InitialPassword)
	}
}

func TestLoadReadsInitialPassword(t *testing.T) {
	t.Setenv("BOPHOTOS_INITIAL_PASSWORD", "correct horse battery staple")

	runtime, err := Load()
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if runtime.InitialPassword != "correct horse battery staple" {
		t.Errorf("InitialPassword = %q, want configured password", runtime.InitialPassword)
	}
}

func TestLoadRejectsInvalidMaxUploadBytes(t *testing.T) {
	t.Setenv("BOPHOTOS_MAX_UPLOAD_BYTES", "large")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want invalid byte count error")
	}
}

func TestLoadRejectsInvalidCookieSecure(t *testing.T) {
	t.Setenv("BOPHOTOS_COOKIE_SECURE", "sometimes")

	if _, err := Load(); err == nil {
		t.Fatal("Load() error = nil, want invalid boolean error")
	}
}
