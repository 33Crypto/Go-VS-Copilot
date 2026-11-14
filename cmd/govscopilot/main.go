package main

import (
	"errors"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/33Crypto/Go-VS-Copilot/internal/provider"
	"github.com/33Crypto/Go-VS-Copilot/internal/server"
)

func main() {
	p, err := provider.FromEnv(os.Getenv)
	if err != nil {
		slog.Error("configuration", "err", err)
		os.Exit(1)
	}
	addr := os.Getenv("GOVS_ADDR")
	if addr == "" {
		addr = ":8080"
	}
	srv := &http.Server{
		Addr:              addr,
		Handler:           server.New(p),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      90 * time.Second,
		IdleTimeout:       120 * time.Second,
	}
	slog.Info("listening", "addr", addr, "provider", p.Name(), "model", p.Model())
	if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server stopped", "err", err)
		os.Exit(1)
	}
}
