// Package config loads runtime configuration.
package config

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/spf13/viper"
)

// LogFormat controls the encoding of log records.
type LogFormat string

const (
	// LogFormatText emits human-readable log records.
	LogFormatText LogFormat = "text"
	// LogFormatJSON emits JSON log records.
	LogFormatJSON LogFormat = "json"
)

// Config contains service runtime settings.
type Config struct {
	HTTPAddr    string
	DatabaseURL string
	JWTSecret   string
	LogFormat   LogFormat
	LogLevel    slog.Level
}

// Load reads configuration from an optional file and COMPANIES_ environment variables.
func Load(path string) (Config, error) {
	loader := viper.New()
	loader.SetEnvPrefix("COMPANIES")
	loader.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	loader.AutomaticEnv()
	loader.SetDefault("log_format", LogFormatText)
	loader.SetDefault("log_level", slog.LevelInfo.String())

	if path != "" {
		loader.SetConfigFile(path)
		if err := loader.ReadInConfig(); err != nil {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}

	logFormat := LogFormat(strings.ToLower(strings.TrimSpace(loader.GetString("log_format"))))
	if logFormat == "" {
		logFormat = LogFormatText
	}
	if logFormat != LogFormatText && logFormat != LogFormatJSON {
		return Config{}, fmt.Errorf("invalid log format %q: must be text or json", logFormat)
	}

	var logLevel slog.Level
	if err := logLevel.UnmarshalText([]byte(loader.GetString("log_level"))); err != nil {
		return Config{}, fmt.Errorf("parse log level: %w", err)
	}

	cfg := Config{
		HTTPAddr:    loader.GetString("http_addr"),
		DatabaseURL: loader.GetString("database_url"),
		JWTSecret:   loader.GetString("jwt_secret"),
		LogFormat:   logFormat,
		LogLevel:    logLevel,
	}

	required := []struct {
		name  string
		value string
	}{
		{
			name:  "COMPANIES_HTTP_ADDR",
			value: cfg.HTTPAddr,
		},
		{
			name:  "COMPANIES_DATABASE_URL",
			value: cfg.DatabaseURL,
		},
		{
			name:  "COMPANIES_JWT_SECRET",
			value: cfg.JWTSecret,
		},
	}
	missing := make([]string, 0, len(required))
	for _, setting := range required {
		if strings.TrimSpace(setting.value) == "" {
			missing = append(missing, setting.name)
		}
	}
	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required configuration: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}
