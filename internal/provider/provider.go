// Package provider completes a coding-assistant turn.
// The local stub needs no API key. An OpenAI-compatible provider is used
// only when a key is present in the environment.
package provider

import "context"

// Message is one earlier turn in the conversation.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Request is one completion. Prompt is required. Context is optional source
// the user pasted beside the prompt. Messages are earlier turns, not including
// the current prompt.
type Request struct {
	Prompt   string    `json:"prompt"`
	Context  string    `json:"context,omitempty"`
	Messages []Message `json:"messages,omitempty"`
}

// Response is the assistant reply and the provider that produced it.
type Response struct {
	Reply    string `json:"reply"`
	Provider string `json:"provider"`
	Model    string `json:"model"`
}

// Provider produces a reply for a completion request.
type Provider interface {
	Name() string
	Model() string
	Complete(ctx context.Context, req Request) (Response, error)
}
