package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestStubEchoesPromptAndContext(t *testing.T) {
	got, err := Stub{}.Complete(context.Background(), Request{
		Prompt:  "explain this loop",
		Context: "for {}",
	})
	if err != nil {
		t.Fatal(err)
	}
	if got.Provider != "stub" || got.Model != "local-echo" {
		t.Fatalf("identity = %+v", got)
	}
	if !strings.Contains(got.Reply, "explain this loop") || !strings.Contains(got.Reply, "for {}") {
		t.Fatalf("reply = %q", got.Reply)
	}
	if !strings.Contains(got.Reply, "No model was called") {
		t.Fatalf("reply = %q", got.Reply)
	}
}

func TestStubOmitsEmptyContext(t *testing.T) {
	got, err := Stub{}.Complete(context.Background(), Request{Prompt: "hi", Context: " \n\t "})
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(got.Reply, "--- context ---") {
		t.Fatalf("reply = %q", got.Reply)
	}
}

func TestFromEnv(t *testing.T) {
	t.Run("no key uses stub", func(t *testing.T) {
		p, err := FromEnv(func(string) string { return "" })
		if err != nil {
			t.Fatal(err)
		}
		if p.Name() != "stub" {
			t.Fatalf("name = %s", p.Name())
		}
	})

	t.Run("key uses openai defaults", func(t *testing.T) {
		p, err := FromEnv(func(k string) string {
			if k == "OPENAI_API_KEY" {
				return "sk-test"
			}
			return ""
		})
		if err != nil {
			t.Fatal(err)
		}
		if p.Name() != "openai" || p.Model() != "gpt-4o-mini" {
			t.Fatalf("provider = %s %s", p.Name(), p.Model())
		}
		o := p.(*openAI)
		if o.baseURL != "https://api.openai.com/v1" {
			t.Fatalf("base = %s", o.baseURL)
		}
		if o.apiKey != "sk-test" {
			t.Fatal("key not stored")
		}
	})

	t.Run("stub wins over a key", func(t *testing.T) {
		p, err := FromEnv(func(k string) string {
			switch k {
			case "GOVS_PROVIDER":
				return "stub"
			case "GOVS_API_KEY":
				return "sk-test"
			default:
				return ""
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		if p.Name() != "stub" {
			t.Fatalf("name = %s", p.Name())
		}
	})

	t.Run("openai without key fails", func(t *testing.T) {
		_, err := FromEnv(func(k string) string {
			if k == "GOVS_PROVIDER" {
				return "openai"
			}
			return ""
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("unknown provider fails", func(t *testing.T) {
		_, err := FromEnv(func(k string) string {
			if k == "GOVS_PROVIDER" {
				return "nope"
			}
			return ""
		})
		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("custom base and model", func(t *testing.T) {
		p, err := FromEnv(func(k string) string {
			switch k {
			case "GOVS_API_KEY":
				return "local-key"
			case "GOVS_BASE_URL":
				return "http://127.0.0.1:11434/v1/"
			case "GOVS_MODEL":
				return "llama3.2"
			default:
				return ""
			}
		})
		if err != nil {
			t.Fatal(err)
		}
		o := p.(*openAI)
		if o.baseURL != "http://127.0.0.1:11434/v1/" || o.model != "llama3.2" {
			t.Fatalf("%+v", o)
		}
	})
}

func TestOpenAIComplete(t *testing.T) {
	const key = "sk-test-secret-value"
	var seenAuth string
	var seen chatRequest
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		seenAuth = r.Header.Get("Authorization")
		if r.URL.Path != "/chat/completions" {
			t.Errorf("path = %s", r.URL.Path)
		}
		raw, _ := io.ReadAll(r.Body)
		if strings.Contains(string(raw), key) {
			t.Errorf("api key appeared in the request body")
		}
		if err := json.Unmarshal(raw, &seen); err != nil {
			t.Errorf("decode request: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"use a range loop"}}]}`)
	}))
	defer srv.Close()

	p := &openAI{apiKey: key, baseURL: srv.URL + "/", model: "test-model", client: srv.Client()}
	got, err := p.Complete(context.Background(), Request{
		Prompt:   "how do I iterate",
		Context:  "items := []int{1}",
		Messages: []Message{{Role: "user", Content: "earlier question"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if seenAuth != "Bearer "+key {
		t.Fatalf("auth = %q", seenAuth)
	}
	if seen.Model != "test-model" {
		t.Fatalf("model = %s", seen.Model)
	}
	blob := mustJSON(t, seen.Messages)
	for _, want := range []string{"how do I iterate", "items := []int{1}", "earlier question", "coding assistant"} {
		if !strings.Contains(blob, want) {
			t.Fatalf("messages missing %q: %s", want, blob)
		}
	}
	if got.Reply != "use a range loop" || got.Provider != "openai" || got.Model != "test-model" {
		t.Fatalf("%+v", got)
	}
}

func TestOpenAIDoesNotLeakAPIKey(t *testing.T) {
	const key = "sk-test-secret-value"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, "bad key "+key)
	}))
	defer srv.Close()

	p := &openAI{apiKey: key, baseURL: srv.URL, model: "test-model", client: srv.Client()}
	_, err := p.Complete(context.Background(), Request{Prompt: "hi"})
	if err == nil {
		t.Fatal("expected error")
	}
	if strings.Contains(err.Error(), key) {
		t.Fatalf("key leaked: %v", err)
	}
	if !strings.Contains(err.Error(), "401") {
		t.Fatalf("err = %v", err)
	}
}

func TestOpenAIEmptyCompletion(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `{"choices":[{"message":{"role":"assistant","content":"  "}}]}`)
	}))
	defer srv.Close()
	p := &openAI{apiKey: "k", baseURL: srv.URL, model: "m", client: srv.Client()}
	_, err := p.Complete(context.Background(), Request{Prompt: "hi"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func mustJSON(t *testing.T, v any) string {
	t.Helper()
	b, err := json.Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}
