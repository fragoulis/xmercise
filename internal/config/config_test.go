package config_test

import (
	"log/slog"
	"testing"

	"github.com/fragoulis/xmercise/internal/config"
)

func TestLoad(t *testing.T) {
	tests := map[string]struct {
		logFormat  string
		logLevel   string
		wantFormat config.LogFormat
		wantLevel  slog.Level
	}{
		"defaults": {
			wantFormat: config.LogFormatText,
			wantLevel:  slog.LevelInfo,
		},
		"loads configured values": {
			logFormat:  "json",
			logLevel:   "debug",
			wantFormat: config.LogFormatJSON,
			wantLevel:  slog.LevelDebug,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("COMPANIES_HTTP_ADDR", ":8080")
			t.Setenv("COMPANIES_DATABASE_URL", "postgres://companies:companies@localhost/companies")
			t.Setenv("COMPANIES_JWT_SECRET", "secret")
			t.Setenv("COMPANIES_LOG_FORMAT", test.logFormat)
			t.Setenv("COMPANIES_LOG_LEVEL", test.logLevel)

			cfg, err := config.Load("")
			if err != nil {
				t.Fatalf("Load() error = %v", err)
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

func TestLoadRejectsInvalidLoggingConfiguration(t *testing.T) {
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
