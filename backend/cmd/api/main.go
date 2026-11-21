// Command api runs the rpgmaker HTTP server.
package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/auth"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/campaigns"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/health"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/migrations"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/platform/database"
	"github.com/joao-gabriel-cruz/rpgmaker/backend/internal/systems"
)

func main() {
	if err := run(); err != nil {
		slog.Error("api exited with error", "err", err)
		os.Exit(1)
	}
}

func run() error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if len(os.Args) > 1 && os.Args[1] == "create-admin" {
		return createAdmin(ctx, os.Args[2:])
	}

	pool, err := setupDatabase(ctx)
	if err != nil {
		return err
	}
	defer pool.Close()

	mux := http.NewServeMux()
	mux.Handle("GET /api/health", health.Handler(pool))
	authSvc := auth.NewService(auth.NewStore(pool))
	authHandlers := auth.NewHandlers(authSvc)
	authHandlers.Register(mux)
	reg := systems.Default()
	systems.Register(mux, reg, authHandlers.RequireAuth)
	campaigns.NewHandlers(campaigns.NewService(campaigns.NewStore(pool), authSvc, reg), authHandlers.RequireAuth).Register(mux)

	addr := ":" + envOr("PORT", "8080")
	return serve(ctx, &http.Server{Addr: addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second})
}

func setupDatabase(ctx context.Context) (*pgxpool.Pool, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return nil, errors.New("DATABASE_URL is not set")
	}
	pool, err := database.Connect(ctx, dbURL)
	if err != nil {
		return nil, err
	}
	if err := migrations.Up(ctx, database.StdDB(pool)); err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// serve blocks until the server fails or ctx is cancelled, then shuts down gracefully.
func serve(ctx context.Context, srv *http.Server) error {
	errCh := make(chan error, 1)
	go func() {
		slog.Info("api listening", "addr", srv.Addr)
		errCh <- srv.ListenAndServe()
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return srv.Shutdown(shutdownCtx)
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
