package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type openAI struct {
	apiKey  string
	baseURL string
	model   string
	client  *http.Client
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatResponse struct {
	Choices []struct {
		Message chatMessage `json:"message"`
	} `json:"choices"`
}

func (c *openAI) Name() string  { return "openai" }
func (c *openAI) Model() string { return c.model }

func (c *openAI) Complete(ctx context.Context, req Request) (Response, error) {
	messages := make([]chatMessage, 0, len(req.Messages)+2)
	messages = append(messages, chatMessage{
		Role:    "system",
		Content: "You are Go-VS-Copilot, a coding assistant. Prefer concise answers. When code context is included, ground the answer in that context.",
	})
	for _, m := range req.Messages {
		messages = append(messages, chatMessage{Role: m.Role, Content: m.Content})
	}
	user := req.Prompt
	if strings.TrimSpace(req.Context) != "" {
		user += "\n\nCode context:\n" + req.Context
	}
	messages = append(messages, chatMessage{Role: "user", Content: user})

	payload, err := json.Marshal(chatRequest{Model: c.model, Messages: messages})
	if err != nil {
		return Response{}, err
	}
	endpoint := strings.TrimRight(c.baseURL, "/") + "/chat/completions"
	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(payload))
	if err != nil {
		return Response{}, c.redact(err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	client := c.client
	if client == nil {
		client = &http.Client{Timeout: 60 * time.Second}
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		return Response{}, c.redact(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return Response{}, c.redact(err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		// The upstream body is discarded from the error. It can echo the key.
		return Response{}, fmt.Errorf("provider status %d", resp.StatusCode)
	}
	var parsed chatResponse
	if err := json.Unmarshal(body, &parsed); err != nil {
		return Response{}, errors.New("provider returned invalid json")
	}
	if len(parsed.Choices) == 0 || strings.TrimSpace(parsed.Choices[0].Message.Content) == "" {
		return Response{}, errors.New("provider returned an empty completion")
	}
	return Response{
		Reply:    parsed.Choices[0].Message.Content,
		Provider: c.Name(),
		Model:    c.model,
	}, nil
}

func (c *openAI) redact(err error) error {
	if err == nil || c.apiKey == "" {
		return err
	}
	return errors.New(strings.ReplaceAll(err.Error(), c.apiKey, "[redacted]"))
}
