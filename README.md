# Go-VS-Copilot

Go-VS-Copilot is a coding-assistant chat built by [LOCK33 Development Organization](https://lock33.dev). A Go server serves a small web app that runs in Chrome on Linux and in Chrome on Android. You type a prompt, optionally paste code, and get a reply.

The page uses the LOCK33 colors and links to [lock33.dev](https://lock33.dev). Those files are part of this app. Nothing is loaded from lock33.dev at runtime, so the installed page still opens offline.

With no API key the server uses a local stub. The stub repeats your prompt and code context and does not call a model. Set the environment variables below when you want an OpenAI-compatible API to answer instead.

## Run

The module targets the current stable Go release, 1.27.

```bash
go test ./...
go run ./cmd/govscopilot
```

Then open [http://127.0.0.1:8080](http://127.0.0.1:8080).

`GOVS_ADDR` changes the listen address. The default is `:8080`, which accepts connections on every interface. Build a binary with:

```bash
go build -o govscopilot ./cmd/govscopilot
./govscopilot
```

## Chrome on Linux

1. Start the server.
2. Open [http://127.0.0.1:8080](http://127.0.0.1:8080) in Chrome.
3. Install it from the install icon in the address bar.

The page includes a web app manifest and a service worker. `localhost` is a secure context, so Chrome can install the app. The installed window opens the same chat.

## Chrome on Android

On a phone that can reach the computer, open `http://<linux-lan-address>:8080`. The chat works there over plain HTTP.

Installing the app on Android (Add to Home screen) needs HTTPS that the phone trusts. A service worker is registered only in a secure context, and `localhost` on the phone is the phone, not the Linux machine. Keep using the page in the browser until the server is behind HTTPS. The manifest and service worker are already part of the page, so the same URL works for install once it is served over a trusted certificate.

The server has no login. On a shared network, anyone who can open the port can send prompts. If an API key is configured, those prompts are sent with that key. Use localhost, or a private network you trust.

## Chat

The prompt box is the question. **Code context** is the optional snippet. Ctrl+Enter or ⌘+Enter sends. On a phone, use the Send button. The pill in the header shows `Local stub` or the remote provider and model.

`POST /api/complete`

```json
{
  "prompt": "why does this loop skip the last item?",
  "context": "for i := 0; i < n; i++ {}",
  "messages": []
}
```

`prompt` is required. `context` is optional source. `messages` is earlier turns (`user` or `assistant`), not including the current prompt. The reply is JSON: `reply`, `provider`, and `model`.

`GET /api/health` returns `{"ok":true,"provider":"stub","model":"local-echo"}` when the stub is active.

## Use a model

No key is stored in this repository. The stub stays active until a key is present.

```bash
export GOVS_API_KEY="the key from your provider"
export GOVS_MODEL="gpt-4o-mini"
export GOVS_BASE_URL="https://api.openai.com/v1"
go run ./cmd/govscopilot
```

- `GOVS_API_KEY`, or `OPENAI_API_KEY` if the first is unset, selects the OpenAI-compatible provider.
- `GOVS_BASE_URL` is the API root. The client calls `{GOVS_BASE_URL}/chat/completions`. The default root is `https://api.openai.com/v1`.
- `GOVS_MODEL` is sent as the model name. The default is `gpt-4o-mini`. Set this to the name your provider expects.
- `GOVS_PROVIDER=stub` forces the local echo even when a key is set.
- `GOVS_PROVIDER=openai` requires a key and exits on startup if it is missing.

A local server that speaks the same chat-completions API can be used by setting `GOVS_BASE_URL` to its root and `GOVS_API_KEY` to whatever that server expects. An empty key always means the stub.

## Layout

- `cmd/govscopilot` starts the server.
- `internal/provider` is the completion interface, the local stub, and the OpenAI-compatible client.
- `internal/server` is the HTTP API and the tests for it.
- `web` is the page, stylesheet, script, manifest, service worker, and icons. They are embedded in the binary.
- `.github/workflows/ci.yml` runs `go test ./...`, `go build`, and `gofmt` on Ubuntu with Go 1.27.3.

## What was removed

This branch no longer contains the Windows Azure Table Storage sample or the template leftovers that shipped with it (`ci.go`, `go.os`, `Index.cs.tar.zst`, `cosx-env.yml`, the CodeQL stub, `scripts/all-cmd.sh`, the empty Word document, and `.github/cosx-erb.yml`). The Azure sample remains in [33Crypto/WindowsAzureRepository](https://github.com/33Crypto/WindowsAzureRepository). That repository is not modified by this project.
