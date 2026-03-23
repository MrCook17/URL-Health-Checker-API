package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"time"

	"healthcheck-api/internal/checker"
	"healthcheck-api/internal/httpapi"
	"healthcheck-api/internal/store"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	st := store.NewMemoryStore()

	client := &http.Client{}
	chk := checker.NewHTTPChecker(client)

	h := httpapi.NewHandler(st, chk)

	mux := http.NewServeMux()
	h.Register(mux)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           httpapi.Middleware(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	serverErr := make(chan error, 1)

	go func() {
		slog.Info("server_starting", "addr", server.Addr)

		err := server.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	select {
	case <-ctx.Done():
		slog.Info("shutdown_signal_received")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			slog.Error("server_shutdown_failed", "error", err)
			os.Exit(1)
		}

		slog.Info("server_stopped")
	case err := <-serverErr:
		if err != nil {
			slog.Error("server_listen_failed", "error", err)
			os.Exit(1)
		}
	}
}