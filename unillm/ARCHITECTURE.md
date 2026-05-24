# UniLLM Architecture

## Overview
Multi-model AI API aggregation platform. OpenAI-compatible proxy that routes requests to upstream providers (OpenAI, Anthropic, Google, DeepSeek, Alibaba, ByteDance).

## Tech Stack
- **Backend**: Go 1.26 + Gin
- **Database**: PostgreSQL 16 + Redis 7
- **Frontend**: Next.js + shadcn/ui (planned)
- **Deploy**: Docker Compose

## Project Structure
```
unillm/
├── cmd/server/main.go              # Entry point, DI wiring
├── internal/
│   ├── config/                      # Environment config
│   ├── handler/
│   │   ├── auth.go                  # Register, login, API key CRUD
│   │   ├── models.go                # GET /v1/models
│   │   └── proxy.go                 # POST /v1/chat/completions (core)
│   ├── middleware/
│   │   ├── auth.go                  # JWT auth + API key auth
│   │   └── ratelimit.go             # Per-user rate limiting
│   ├── model/models.go              # GORM models
│   ├── provider/
│   │   ├── provider.go              # Provider interface + Registry
│   │   ├── openai_provider.go       # OpenAI-compatible adapter
│   │   ├── anthropic_provider.go    # Anthropic Messages → OpenAI translation
│   │   └── google_provider.go       # Gemini → OpenAI translation
│   ├── repository/                  # Data access layer
│   └── service/
│       ├── auth.go                  # Auth + API key management
│       └── billing.go               # Redis hot path + PG flush worker
├── pkg/openai/types.go              # OpenAI request/response types
├── migrations/001_init.sql          # Database schema
├── docker-compose.yml
└── Dockerfile
```

## API Routes
- `POST /api/auth/register` — User registration
- `POST /api/auth/login` — JWT login
- `GET /api/me` — User profile (JWT)
- `POST /api/keys` — Create API key (JWT)
- `GET /api/keys` — List API keys (JWT)
- `DELETE /api/keys` — Delete API key (JWT)
- `GET /v1/models` — List available models (API key)
- `POST /v1/chat/completions` — Chat proxy (API key, streaming + non-streaming)

## Key Decisions
1. **Provider interface pattern**: All providers implement `Provider` interface with `ChatCompletion` and `ChatCompletionStream`
2. **OpenAI-compatible format**: External API is 100% OpenAI-compatible. Anthropic/Gemini translation in adapters
3. **API key hashing**: SHA-256 hash stored in DB, raw key shown only once at creation
4. **Round-robin key pool**: Multiple upstream keys per provider for load distribution
5. **Zero model falsification**: Only models with active config are listed in /v1/models
6. **Redis billing buffer**: Usage counters in Redis (atomic increment), flush to PG every 5s
7. **Anthropic SSE adapter**: Stream adapter converts Anthropic SSE events to OpenAI chunk format in real-time

## Verified Working (with evidence)
- User registration + JWT login ✅
- API key creation + SHA-256 hashing ✅
- `/v1/models` returns 3 configured models ✅
- DeepSeek proxy: UniLLM → Geneasy → DeepSeek (6.7s, 63 tokens) ✅
- Claude Haiku proxy: UniLLM → Geneasy → Claude (4.6s, 64 tokens) ✅
- Gemini Flash proxy: UniLLM → Geneasy → Gemini (1.1s, 22 tokens) ✅
- Redis billing counters (per-user daily, per-model hourly) ✅
- PG usage log flush (3 rows visible in usage_logs table) ✅
- Rate limiting middleware (200 req/min/user) ✅

## TODO (Next Sessions)
- [ ] Streaming proxy test (SSE end-to-end)
- [ ] Anthropic SSE stream adapter testing with real API
- [ ] Usage stats API (dashboard endpoints)
- [ ] Next.js frontend dashboard
- [ ] Status page (model health monitoring)
- [ ] Balance check middleware (reject if insufficient)
- [ ] Admin API (manage providers, models, users)
