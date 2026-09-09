package config_test

import (
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/fragoulis/xmercise/internal/config"
)

func TestLoad(t *testing.T) {
	tests := map[string]struct {
		authEnabled string
		logFormat   string
		logLevel    string
		wantAuth    bool
		wantFormat  config.LogFormat
		wantLevel   slog.Level
	}{
		"defaults": {
			wantAuth:   true,
			wantFormat: config.LogFormatText,
			wantLevel:  slog.LevelInfo,
		},
		"loads configured values": {
			authEnabled: "false",
			logFormat:   "json",
			logLevel:    "debug",
			wantFormat:  config.LogFormatJSON,
			wantLevel:   slog.LevelDebug,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("COMPANIES_HTTP_ADDR", ":8080")
			t.Setenv("COMPANIES_DATABASE_URL", "postgres://companies:companies@localhost/companies")
			t.Setenv("COMPANIES_JWT_SECRET", "secret")
			t.Setenv("COMPANIES_AUTH_ENABLED", test.authEnabled)
			t.Setenv("COMPANIES_LOG_FORMAT", test.logFormat)
			t.Setenv("COMPANIES_LOG_LEVEL", test.logLevel)

			cfg, err := config.Load("")
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.AuthEnabled != test.wantAuth {
				t.Errorf("AuthEnabled = %t, want %t", cfg.AuthEnabled, test.wantAuth)
			}
			if cfg.LogFormat != test.wantFormat {
				t.Errorf("LogFormat = %q, want %q", cfg.LogFormat, test.wantFormat)
			}
			if cfg.LogLevel != test.wantLevel {
				t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, test.wantLevel)
			}
		})
	}
}

func TestLoadAllowsDisabledAuthenticationWithoutJWTSecret(t *testing.T) {
	t.Setenv("COMPANIES_HTTP_ADDR", ":8080")
	t.Setenv("COMPANIES_DATABASE_URL", "postgres://companies:companies@localhost/companies")
	t.Setenv("COMPANIES_AUTH_ENABLED", "false")

	cfg, err := config.Load("")
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AuthEnabled {
		t.Error("AuthEnabled = true, want false")
	}
}

func TestLoadFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	contents := []byte(`
http_addr: ":8080"
database_url: "postgres://companies:companies@localhost/companies"
auth_enabled: false
log_level: debug
`)
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	cfg, err := config.Load(path)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.AuthEnabled {
		t.Error("AuthEnabled = true, want false")
	}
	if cfg.LogLevel != slog.LevelDebug {
		t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, slog.LevelDebug)
	}
}

func TestLoadRejectsInvalidConfiguration(t *testing.T) {
	tests := map[string]struct {
		name  string
		value string
	}{
		"format": {
			name:  "COMPANIES_LOG_FORMAT",
			value: "csv",
		},
		"level": {
			name:  "COMPANIES_LOG_LEVEL",
			value: "verbose",
		},
		"authentication": {
			name:  "COMPANIES_AUTH_ENABLED",
			value: "sometimes",
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("COMPANIES_HTTP_ADDR", ":8080")
			t.Setenv("COMPANIES_DATABASE_URL", "postgres://companies:companies@localhost/companies")
			t.Setenv("COMPANIES_JWT_SECRET", "secret")
			t.Setenv(test.name, test.value)

			_, err := config.Load("")
			if err == nil {
				t.Fatal("Load() error = nil, want invalid logging configuration error")
			}
		})
	}
}
