package provider

import (
	"fmt"
	"strings"
)

// FromEnv selects a provider from the process environment.
//
// GOVS_PROVIDER=stub forces the local echo provider.
// GOVS_PROVIDER=openai requires GOVS_API_KEY or OPENAI_API_KEY.
// When GOVS_PROVIDER is empty, a key selects the OpenAI-compatible provider
// and a missing key selects the stub.
//
// GOVS_BASE_URL is the API root (default https://api.openai.com/v1).
// GOVS_MODEL is the model name (default gpt-4o-mini).
func FromEnv(getenv func(string) string) (Provider, error) {
	if getenv == nil {
		return nil, fmt.Errorf("getenv is nil")
	}
	mode := strings.ToLower(strings.TrimSpace(getenv("GOVS_PROVIDER")))
	key := firstNonEmpty(getenv("GOVS_API_KEY"), getenv("OPENAI_API_KEY"))
	switch mode {
	case "", "auto":
		if key == "" {
			return Stub{}, nil
		}
		return newOpenAI(getenv, key), nil
	case "stub", "local", "echo":
		return Stub{}, nil
	case "openai":
		if strings.TrimSpace(key) == "" {
			return nil, fmt.Errorf("GOVS_PROVIDER=openai requires GOVS_API_KEY or OPENAI_API_KEY")
		}
		return newOpenAI(getenv, key), nil
	default:
		return nil, fmt.Errorf("unknown GOVS_PROVIDER %q", mode)
	}
}

func newOpenAI(getenv func(string) string, key string) Provider {
	base := strings.TrimSpace(getenv("GOVS_BASE_URL"))
	if base == "" {
		base = "https://api.openai.com/v1"
	}
	model := strings.TrimSpace(getenv("GOVS_MODEL"))
	if model == "" {
		model = "gpt-4o-mini"
	}
	return &openAI{apiKey: key, baseURL: base, model: model}
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}
