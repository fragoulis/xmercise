package config_test

import (
	"log/slog"
	"testing"

	"github.com/fragoulis/xmercise/internal/config"
)

func TestLoad(t *testing.T) {
	tests := map[string]struct {
		logLevel string
		want     slog.Level
	}{
		"defaults to info": {
			want: slog.LevelInfo,
		},
		"loads configured level": {
			logLevel: "debug",
			want:     slog.LevelDebug,
		},
	}

	for name, test := range tests {
		t.Run(name, func(t *testing.T) {
			t.Setenv("COMPANIES_HTTP_ADDR", ":8080")
			t.Setenv("COMPANIES_DATABASE_URL", "postgres://companies:companies@localhost/companies")
			t.Setenv("COMPANIES_JWT_SECRET", "secret")
			t.Setenv("COMPANIES_LOG_LEVEL", test.logLevel)

			cfg, err := config.Load("")
			if err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if cfg.LogLevel != test.want {
				t.Errorf("LogLevel = %v, want %v", cfg.LogLevel, test.want)
			}
		})
	}
}

func TestLoadRejectsInvalidLogLevel(t *testing.T) {
	t.Setenv("COMPANIES_HTTP_ADDR", ":8080")
	t.Setenv("COMPANIES_DATABASE_URL", "postgres://companies:companies@localhost/companies")
	t.Setenv("COMPANIES_JWT_SECRET", "secret")
	t.Setenv("COMPANIES_LOG_LEVEL", "verbose")

	_, err := config.Load("")
	if err == nil {
		t.Fatal("Load() error = nil, want invalid log level error")
	}
}
