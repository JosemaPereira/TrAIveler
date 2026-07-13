# Local AI Setup (Ollama + Gemma)

## Why

`docs/architecture.md` and `docs/cloud-and-environments.md` document **Anthropic Claude** as the
AI provider for staging and production. For **local development and MVP testing**, TrAIveler uses
**[Ollama](https://ollama.com) running a local Gemma model** instead — it's free, requires no API
key, and needs no network access once the model is downloaded. This is a local-development
substitution only; it does not change the product's staging/production architecture (see the
"Local Development Note" addenda in `docs/architecture.md` and `docs/cloud-and-environments.md`).

`backend/internal/ai/` implements the `AIClient` interface against both backends:
`OllamaClient` (`ollama_client.go`, used here) and `AnthropicClient` (`anthropic.go`,
staging/production). `NewAIClient` (`client.go`) is the factory that picks between them based on
the `AI_PROVIDER` environment variable — see `backend/config/config.go`.

## Prerequisites

- macOS, Linux, or Windows (WSL2)
- ~4 GB free disk space for the model, ~4 GB free RAM to run it comfortably

## 1. Install Ollama

Download and install from [ollama.com/download](https://ollama.com/download), or via a package
manager:

```bash
# macOS (Homebrew)
brew install ollama

# Linux
curl -fsSL https://ollama.com/install.sh | sh
```

Start the server (if it isn't already running as a background service after install):

```bash
ollama serve
```

By default it listens on `http://localhost:11434`.

## 2. Pull a Gemma model

```bash
ollama pull gemma3:4b
```

`gemma3:4b` is the default configured in `backend/.env.example` (`OLLAMA_MODEL`) — a reasonable
balance of quality and resource usage for itinerary-generation testing. If your machine is
resource-constrained, use a smaller variant instead and set `OLLAMA_MODEL` to match:

| Model tag     | Approx. size | When to use                                  |
|---------------|-------------:|-----------------------------------------------|
| `gemma3:1b`   | ~1 GB        | Low-RAM machines, fastest responses           |
| `gemma3:4b`   | ~3 GB        | Default — good quality/speed balance          |
| `gemma3:12b`  | ~8 GB        | Higher quality, needs a beefier machine       |

## 3. Verify the server is reachable

```bash
curl http://localhost:11434/api/tags
```

You should get back a JSON list including the model(s) you pulled. If this fails, confirm
`ollama serve` is running and nothing else is bound to port `11434`.

## 4. Configure the backend

`backend/.env.example` already defaults to Ollama — copy it if you haven't yet:

```bash
cp backend/.env.example backend/.env
```

Relevant variables (see `backend/.env.example` for the full block):

| Variable         | Default                    | Notes                                              |
|------------------|-----------------------------|-----------------------------------------------------|
| `GO_ENV`         | `development`               | `dev`/`prod` above means `GO_ENV != production` vs. `GO_ENV == production` — it's the switch behind `AI_PROVIDER`'s default |
| `AI_PROVIDER`    | `ollama` (dev) / `anthropic` (prod) | Selects the backend; `config.Load()` fails fast on any other value |
| `OLLAMA_HOST`    | `http://localhost:11434`   | Set to `http://ollama:11434` when running via `docker-compose` (Docker DNS hostname) |
| `OLLAMA_MODEL`   | `gemma3:4b`                | Must match a model you've pulled                    |
| `AI_TIMEOUT`     | `60s`                       | Shared across providers                              |
| `AI_MAX_RETRIES` | `3`                         | Retries on connection errors / 5xx from the provider |

No `ANTHROPIC_API_KEY` is required with the default `AI_PROVIDER=ollama`.

## 5. Run it

**Option A — Docker Compose (recommended, matches CI/staging topology):**

```bash
docker-compose up -d
docker-compose exec ollama ollama pull gemma3:4b   # one-time, if the volume is empty
```

This starts Postgres, Ollama, and the backend together on the same Docker network — the backend
reaches Ollama at `http://ollama:11434` automatically (already set in `docker-compose.yml`).

**Option B — Native Ollama + backend running on the host:**

Run `ollama serve` natively (steps above), then run the backend as usual (`go run ./cmd/api` from
`backend/`, or via your existing local Postgres setup) with `backend/.env`'s default
`OLLAMA_HOST=http://localhost:11434`.

## 6. Exercise the AI client directly

Trip-generation endpoints (spec 008) aren't wired into the HTTP API yet — this ticket only
establishes the `AIClient` interface, stubs, and the concrete `OllamaClient`. To confirm your local
setup actually works end-to-end today, either:

- Run the package's own test suite, which exercises the real wire format against a fake server:
  `cd backend && go test -tags=test ./internal/ai/...`
- Or call Ollama's API directly to confirm the model responds in the JSON-mode `OllamaClient` relies
  on:

  ```bash
  curl http://localhost:11434/api/chat -d '{
    "model": "gemma3:4b",
    "format": "json",
    "stream": false,
    "messages": [{"role": "user", "content": "Reply with {\"ok\": true}"}]
  }'
  ```

## Switching to Anthropic (staging/production parity testing)

Set `AI_PROVIDER=anthropic` and provide `ANTHROPIC_API_KEY` (see
[console.anthropic.com/settings/keys](https://console.anthropic.com/settings/keys)) and, optionally,
`ANTHROPIC_MODEL`. `config.Load()` will then require the key, and `NewAIClient` (`client.go`) will
route requests through `AnthropicClient` (`anthropic.go`) instead of `OllamaClient` — see
`backend/README.md`'s "AI client foundation" section for how retries/timeouts and provider-outage
handling (`Retry-After`) work for that path.

## Troubleshooting

- **`connection refused` on `:11434`**: `ollama serve` isn't running, or you're using
  `docker-compose` but `OLLAMA_HOST` still points at `localhost` instead of `http://ollama:11434`.
- **Model responses are slow or the container OOMs**: drop to a smaller model tag (`gemma3:1b`) or
  raise the `ollama` service's resource limits in `docker-compose.yml`.
- **`ollama: parse itinerary JSON content` errors**: the model didn't return content matching
  `ai.ItineraryResponse`'s shape. Expected until spec 008 adds real system-prompt/schema guidance —
  `OllamaClient` forwards `ConversationHistory` as-is today (see `ollama_client.go`).
