// Package config loads runtime configuration.
package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config contains service runtime settings.
type Config struct {
	HTTPAddr    string
	DatabaseURL string
	JWTSecret   string
	JWTIssuer   string
	JWTAudience string
}

// Load reads configuration from an optional file and COMPANIES_ environment variables.
func Load(path string) (Config, error) {
	loader := viper.New()
	loader.SetEnvPrefix("COMPANIES")
	loader.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	loader.AutomaticEnv()

	if path != "" {
		loader.SetConfigFile(path)
		if err := loader.ReadInConfig(); err != nil {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}

	cfg := Config{
		HTTPAddr:    loader.GetString("http_addr"),
		DatabaseURL: loader.GetString("database_url"),
		JWTSecret:   loader.GetString("jwt_secret"),
		JWTIssuer:   loader.GetString("jwt_issuer"),
		JWTAudience: loader.GetString("jwt_audience"),
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
