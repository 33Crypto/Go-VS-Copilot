package server

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/33Crypto/Go-VS-Copilot/internal/provider"
)

func TestHealth(t *testing.T) {
	rec := httptest.NewRecorder()
	New(provider.Stub{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/health", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{`"ok":true`, `"provider":"stub"`, `"model":"local-echo"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("body %s missing %s", body, want)
		}
	}
}

func TestCompleteEchoesPromptAndContext(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/complete", strings.NewReader(`{
		"prompt": "  explain this loop  ",
		"context": "for i := 0; i < n; i++ {}",
		"messages": [{"role":"user","content":"earlier"}]
	}`))
	rec := httptest.NewRecorder()
	New(provider.Stub{}).ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d body %s", rec.Code, rec.Body.String())
	}
	var body struct {
		Reply    string `json:"reply"`
		Provider string `json:"provider"`
		Model    string `json:"model"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Provider != "stub" || body.Model != "local-echo" {
		t.Fatalf("%+v", body)
	}
	for _, want := range []string{"explain this loop", "for i := 0; i < n; i++ {}", "No model was called"} {
		if !strings.Contains(body.Reply, want) {
			t.Fatalf("reply missing %q:\n%s", want, body.Reply)
		}
	}
	if strings.Contains(body.Reply, "  explain this loop  ") {
		t.Fatalf("prompt was not trimmed: %s", body.Reply)
	}
}

func TestCompleteRejectsBadInput(t *testing.T) {
	cases := []struct {
		name string
		body string
		want int
	}{
		{"empty prompt", `{"prompt":"   "}`, http.StatusBadRequest},
		{"missing prompt", `{}`, http.StatusBadRequest},
		{"invalid json", `{`, http.StatusBadRequest},
		{"two values", `{"prompt":"hi"}{"prompt":"again"}`, http.StatusBadRequest},
		{"bad role", `{"prompt":"hi","messages":[{"role":"system","content":"x"}]}`, http.StatusBadRequest},
	}
	h := New(provider.Stub{})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/complete", strings.NewReader(tc.body)))
			if rec.Code != tc.want {
				t.Fatalf("status = %d body %s", rec.Code, rec.Body.String())
			}
			if !strings.Contains(rec.Body.String(), `"error"`) {
				t.Fatalf("body = %s", rec.Body.String())
			}
		})
	}
}

func TestCompleteRejectsLargeBody(t *testing.T) {
	body := `{"prompt":"` + strings.Repeat("a", 1<<20) + `"}`
	rec := httptest.NewRecorder()
	New(provider.Stub{}).ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/complete", strings.NewReader(body)))
	if rec.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d body %s", rec.Code, rec.Body.String())
	}
}

func TestCompleteHidesProviderErrors(t *testing.T) {
	h := New(leakProvider{})
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, httptest.NewRequest(http.MethodPost, "/api/complete", strings.NewReader(`{"prompt":"hi"}`)))
	if rec.Code != http.StatusBadGateway {
		t.Fatalf("status = %d", rec.Code)
	}
	if strings.Contains(rec.Body.String(), "sk-secret") {
		t.Fatalf("provider error leaked: %s", rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), "completion failed") {
		t.Fatalf("body = %s", rec.Body.String())
	}
}

func TestStaticUI(t *testing.T) {
	h := New(provider.Stub{})
	cases := []struct {
		path string
		want string
		typ  string
	}{
		{"/", "Go-VS-Copilot", "text/html"},
		{"/manifest.webmanifest", "standalone", "application/manifest+json"},
		{"/sw.js", `addEventListener("fetch"`, "javascript"},
		{"/app.js", "/api/complete", "javascript"},
		{"/app.css", "--accent", "text/css"},
		{"/icons/icon-192.png", "", "image/png"},
		{"/icons/icon-512.png", "", "image/png"},
	}
	for _, tc := range cases {
		t.Run(tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			h.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d", rec.Code)
			}
			if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, tc.typ) {
				t.Fatalf("content-type = %q", ct)
			}
			if tc.want != "" && !strings.Contains(rec.Body.String(), tc.want) {
				t.Fatalf("body missing %q", tc.want)
			}
			if rec.Header().Get("Content-Security-Policy") == "" {
				t.Fatal("missing content security policy")
			}
		})
	}
}

func TestLock33Branding(t *testing.T) {
	h := New(provider.Stub{})

	page := httptest.NewRecorder()
	h.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/", nil))
	body := page.Body.String()
	for _, want := range []string{
		`href="https://lock33.dev"`,
		`target="_blank"`,
		`rel="noopener noreferrer"`,
		"LOCK33",
		"lock33.dev",
		"Software that stays locked in and holds up.",
		`content="#0b0e13"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("page missing %q", want)
		}
	}
	if strings.Contains(body, "fonts.googleapis") || strings.Contains(body, "lock33.dev/static") {
		t.Fatalf("page loads a remote asset: %s", body)
	}

	css := httptest.NewRecorder()
	h.ServeHTTP(css, httptest.NewRequest(http.MethodGet, "/app.css", nil))
	sheet := css.Body.String()
	for _, want := range []string{"#0b0e13", "#121821", "#263040", "#eef2f6", "#b4bfcc", "#c8ff3e", "#ff9b8f"} {
		if !strings.Contains(sheet, want) {
			t.Fatalf("css missing %s", want)
		}
	}
	if strings.Contains(sheet, "url(http") {
		t.Fatalf("css loads a remote asset")
	}

	manifest := httptest.NewRecorder()
	h.ServeHTTP(manifest, httptest.NewRequest(http.MethodGet, "/manifest.webmanifest", nil))
	if !strings.Contains(manifest.Body.String(), `"theme_color": "#0b0e13"`) {
		t.Fatalf("manifest = %s", manifest.Body.String())
	}

	csp := page.Header().Get("Content-Security-Policy")
	if strings.Contains(csp, "lock33.dev") || strings.Contains(csp, "fonts.g") {
		t.Fatalf("csp allows a third party: %s", csp)
	}
}

func TestIconsDirectoryIsNotListed(t *testing.T) {
	rec := httptest.NewRecorder()
	New(provider.Stub{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/icons/", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("status = %d", rec.Code)
	}
}

func TestCompleteMethodNotAllowed(t *testing.T) {
	rec := httptest.NewRecorder()
	New(provider.Stub{}).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/complete", nil))
	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d", rec.Code)
	}
}

type leakProvider struct{}

func (leakProvider) Name() string  { return "leak" }
func (leakProvider) Model() string { return "x" }
func (leakProvider) Complete(context.Context, provider.Request) (provider.Response, error) {
	return provider.Response{}, errors.New("upstream said sk-secret")
}
