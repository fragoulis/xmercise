package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"

	appcompany "github.com/fragoulis/xmercise/internal/application/company"
	"github.com/fragoulis/xmercise/internal/config"
	"github.com/fragoulis/xmercise/internal/infrastructure/postgres"
	httpadapter "github.com/fragoulis/xmercise/internal/port/http"
)

const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 10 * time.Second
	writeTimeout      = 15 * time.Second
	idleTimeout       = 60 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	command := newCommand()
	command.SetContext(ctx)
	if err := command.Execute(); err != nil {
		slog.Error("service stopped", "error", err)
		os.Exit(1)
	}
}

func newCommand() *cobra.Command {
	var configPath string
	command := &cobra.Command{
		Use:          "companies",
		Short:        "Run the companies service",
		SilenceUsage: true,
		RunE: func(command *cobra.Command, _ []string) error {
			cfg, err := config.Load(configPath)
			if err != nil {
				return err
			}

			return run(command.Context(), cfg)
		},
	}
	command.Flags().StringVar(&configPath, "config", "", "optional configuration file")
	return command
}

func run(ctx context.Context, cfg config.Config) error {
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("create database pool: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}

	store := postgres.NewStore(pool)
	companies := appcompany.NewService(store)
	strictHandler := httpadapter.NewStrictHandler(httpadapter.NewServer(companies), nil)
	auth := httpadapter.NewJWTMiddleware(httpadapter.JWTConfig{
		Secret:   cfg.JWTSecret,
		Issuer:   cfg.JWTIssuer,
		Audience: cfg.JWTAudience,
	})
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           auth.Handler(httpadapter.Handler(strictHandler)),
		ReadHeaderTimeout: readHeaderTimeout,
		ReadTimeout:       readTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}

	serverErrors := make(chan error, 1)
	go func() {
		slog.Info("HTTP server listening", "address", cfg.HTTPAddr)
		serverErrors <- server.ListenAndServe()
	}()

	select {
	case err := <-serverErrors:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return fmt.Errorf("serve HTTP: %w", err)
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shutdown HTTP server: %w", err)
		}
		return nil
	}
}
