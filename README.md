# provider-pin

A tiny loopback proxy that pins OpenRouter requests to one provider.

Some clients (the Codex CLI, for one) cannot send OpenRouter's `provider` routing field. Point them at this proxy and it adds `{"only": ["<provider>"], "allow_fallbacks": false}` to every request whose model starts with a prefix (`deepseek/` by default), then forwards to `https://openrouter.ai`. Everything else passes through untouched. Your API key travels in the request as usual; the proxy never stores or logs it.

## Install

```bash
go install github.com/zbedforrest/provider-pin@latest
```

## Use

DeepSeek V4.1 Flash 1017 on Fireworks, from Codex:

```bash
provider-pin -listen 127.0.0.1:8798 -provider fireworks &

export OPENROUTER_API_KEY=sk-or-...
codex -c model_provider=openrouter \
  -c 'model_providers.openrouter.name="OpenRouter"' \
  -c 'model_providers.openrouter.base_url="http://127.0.0.1:8798/api/v1"' \
  -c 'model_providers.openrouter.env_key="OPENROUTER_API_KEY"' \
  -c 'model_providers.openrouter.wire_api="responses"' \
  -c 'model_reasoning_effort="xhigh"' \
  -m deepseek/deepseek-v4.1-flash-1017
```

Any OpenAI-compatible client works the same way: use `http://127.0.0.1:8798/api/v1` as the base URL.

## Flags

| Flag | Default | Meaning |
|---|---|---|
| `-listen` | `127.0.0.1:8790` | address to listen on |
| `-provider` | (required) | OpenRouter provider slug to pin, e.g. `fireworks`, `digitalocean`, `together` |
| `-match` | `deepseek/` | pin only models whose id starts with this prefix |

`GET /stats` returns how many requests were pinned and how many passed through: `{"passed":0,"pinned":3}`.

With fallbacks off, a request fails if the pinned provider cannot serve the model. Upstream errors (HTTP 400 and above) are logged with their path.
