// Package server is the HTTP API and the Chrome UI.
package server

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/33Crypto/Go-VS-Copilot/internal/provider"
	"github.com/33Crypto/Go-VS-Copilot/web"
)

const (
	maxBody     = 1 << 20
	maxPrompt   = 32_000
	maxContext  = 200_000
	maxMessages = 40
	maxMessage  = 32_000
)

// New returns the application handler. p must be non-nil.
func New(p provider.Provider) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"ok":       true,
			"provider": p.Name(),
			"model":    p.Model(),
		})
	})
	mux.HandleFunc("POST /api/complete", func(w http.ResponseWriter, r *http.Request) {
		complete(w, r, p)
	})
	mux.HandleFunc("GET /{$}", serveEmbedded("index.html", "text/html; charset=utf-8"))
	mux.HandleFunc("GET /app.css", serveEmbedded("app.css", "text/css; charset=utf-8"))
	mux.HandleFunc("GET /app.js", serveEmbedded("app.js", "application/javascript; charset=utf-8"))
	mux.HandleFunc("GET /sw.js", serveEmbedded("sw.js", "application/javascript; charset=utf-8"))
	mux.HandleFunc("GET /manifest.webmanifest", serveEmbedded("manifest.webmanifest", "application/manifest+json; charset=utf-8"))
	mux.HandleFunc("GET /icons/icon-192.png", serveEmbedded("icons/icon-192.png", "image/png"))
	mux.HandleFunc("GET /icons/icon-512.png", serveEmbedded("icons/icon-512.png", "image/png"))

	return withSecurityHeaders(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/complete" && r.Method != http.MethodPost {
			w.Header().Set("Allow", http.MethodPost)
			writeJSON(w, http.StatusMethodNotAllowed, errBody("method not allowed"))
			return
		}
		mux.ServeHTTP(w, r)
	}))
}

func complete(w http.ResponseWriter, r *http.Request, p provider.Provider) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	defer r.Body.Close()

	var req provider.Request
	dec := json.NewDecoder(r.Body)
	if err := dec.Decode(&req); err != nil {
		status := http.StatusBadRequest
		msg := "invalid json"
		var maxErr *http.MaxBytesError
		if errors.As(err, &maxErr) || strings.Contains(err.Error(), "request body too large") {
			status = http.StatusRequestEntityTooLarge
			msg = "request body is too large"
		}
		writeJSON(w, status, errBody(msg))
		return
	}
	if err := dec.Decode(&struct{}{}); err != io.EOF {
		writeJSON(w, http.StatusBadRequest, errBody("invalid json"))
		return
	}

	req.Prompt = strings.TrimSpace(req.Prompt)
	if err := validate(req); err != nil {
		writeJSON(w, http.StatusBadRequest, errBody(err.Error()))
		return
	}

	resp, err := p.Complete(r.Context(), req)
	if err != nil {
		slog.Error("completion failed", "err", err, "provider", p.Name())
		writeJSON(w, http.StatusBadGateway, errBody("completion failed"))
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func validate(req provider.Request) error {
	if req.Prompt == "" {
		return errors.New("prompt is required")
	}
	if utf8.RuneCountInString(req.Prompt) > maxPrompt {
		return errors.New("prompt is too long")
	}
	if utf8.RuneCountInString(req.Context) > maxContext {
		return errors.New("context is too long")
	}
	if len(req.Messages) > maxMessages {
		return errors.New("too many messages")
	}
	for _, m := range req.Messages {
		switch m.Role {
		case "user", "assistant":
		default:
			return errors.New("message role must be user or assistant")
		}
		if utf8.RuneCountInString(m.Content) > maxMessage {
			return errors.New("message is too long")
		}
	}
	return nil
}

func serveEmbedded(name, contentType string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		f, err := web.Files.Open(name)
		if err != nil {
			http.NotFound(w, r)
			return
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil || stat.IsDir() {
			http.NotFound(w, r)
			return
		}
		rs, ok := f.(io.ReadSeeker)
		if !ok {
			http.Error(w, "file is not seekable", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", contentType)
		if name == "sw.js" || name == "app.js" {
			w.Header().Set("Cache-Control", "no-cache")
		}
		http.ServeContent(w, r, name, stat.ModTime(), rs)
	}
}

func withSecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Content-Security-Policy", "default-src 'self'; style-src 'self'; script-src 'self'; img-src 'self'; manifest-src 'self'; worker-src 'self'")
		next.ServeHTTP(w, r)
	})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(v); err != nil {
		slog.Error("write json", "err", err)
	}
}

func errBody(msg string) map[string]string {
	return map[string]string{"error": msg}
}
