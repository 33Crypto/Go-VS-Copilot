package provider

import (
	"context"
	"strings"
)

// Stub is the built-in provider. It echoes the prompt and any code context
// and does not call a network service.
type Stub struct{}

// Name reports the provider id used by the UI and /api/health.
func (Stub) Name() string { return "stub" }

// Model reports the local model id.
func (Stub) Model() string { return "local-echo" }

// Complete returns a deterministic echo of the request.
func (Stub) Complete(ctx context.Context, req Request) (Response, error) {
	if err := ctx.Err(); err != nil {
		return Response{}, err
	}
	var b strings.Builder
	b.WriteString("Local stub. No model was called.\n\n")
	b.WriteString(req.Prompt)
	if strings.TrimSpace(req.Context) != "" {
		b.WriteString("\n\n--- context ---\n")
		b.WriteString(req.Context)
	}
	return Response{
		Reply:    b.String(),
		Provider: "stub",
		Model:    "local-echo",
	}, nil
}
