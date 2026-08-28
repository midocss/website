// Command api runs the monolith HTTP server (public site, store and admin API).
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	"github.com/midocss/website/internal/auth"
	"github.com/midocss/website/internal/catalog"
	"github.com/midocss/website/internal/config"
	"github.com/midocss/website/internal/platform/database"
	"github.com/midocss/website/internal/platform/logger"
	"github.com/midocss/website/internal/rbac"
	transporthttp "github.com/midocss/website/internal/transport/http"
	"github.com/midocss/website/internal/transport/http/handler"
	"github.com/midocss/website/internal/users"
)

// version is overridden at build time with -ldflags "-X main.version=...".
var version = "dev"

func main() {
	if err := run(); err != nil {
		slog.Error("fatal error", "error", err)
		os.Exit(1)
	}
}

func run() error {
	// A missing .env is fine: in production the values come from the environment.
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		return err
	}

	log := logger.New(cfg.App.Environment, cfg.App.Debug)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	db, err := database.Open(ctx, cfg.Database, cfg.App.Debug)
	if err != nil {
		return err
	}
	defer func() {
		if err := database.Close(db); err != nil {
			log.Error("closing database", "error", err)
		}
	}()

	tokens := auth.NewTokenManager(cfg.JWT)
	hasher := auth.NewPasswordHasher(auth.BcryptCost)
	authorizer := rbac.New(db)

	authService := auth.NewService(auth.NewRepository(db), tokens, hasher, authorizer)
	userService := users.NewService(users.NewRepository(db), hasher, authorizer)
	catalogService := catalog.NewService(catalog.NewRepository(db))

	router := transporthttp.NewRouter(transporthttp.Dependencies{
		Config:     cfg,
		Logger:     log,
		Tokens:     tokens,
		Authorizer: authorizer,
		Handlers: transporthttp.Handlers{
			Health:  handler.NewHealthHandler(db, version),
			Auth:    handler.NewAuthHandler(authService),
			Users:   handler.NewUserHandler(userService),
			Catalog: handler.NewCatalogHandler(catalogService),
		},
	})

	server := transporthttp.NewServer(cfg.HTTP, router)

	serverErr := make(chan error, 1)
	go func() {
		log.Info("http server listening", "addr", server.Addr, "version", version)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		log.Info("shutdown signal received")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.HTTP.ShutdownTimeout)
	defer cancel()
	return server.Shutdown(shutdownCtx)
}
