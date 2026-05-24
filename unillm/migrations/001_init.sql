-- UniLLM Initial Schema
-- Designed for multi-model AI API aggregation platform

CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    name VARCHAR(100) NOT NULL DEFAULT '',
    role VARCHAR(20) NOT NULL DEFAULT 'user',
    balance NUMERIC(12, 6) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS api_keys (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL DEFAULT '',
    key_hash VARCHAR(64) UNIQUE NOT NULL,
    key_prefix VARCHAR(12) NOT NULL,
    scope VARCHAR(20) NOT NULL DEFAULT 'full',
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    last_used TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_api_keys_user_id ON api_keys(user_id);

CREATE TABLE IF NOT EXISTS providers (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    base_url VARCHAR(500) NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS provider_keys (
    id BIGSERIAL PRIMARY KEY,
    provider_id BIGINT NOT NULL REFERENCES providers(id) ON DELETE CASCADE,
    key_value TEXT NOT NULL,
    weight INT NOT NULL DEFAULT 1,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    rpm INT NOT NULL DEFAULT 60,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_provider_keys_provider_id ON provider_keys(provider_id);

CREATE TABLE IF NOT EXISTS model_configs (
    id BIGSERIAL PRIMARY KEY,
    public_name VARCHAR(100) UNIQUE NOT NULL,
    provider_id BIGINT NOT NULL REFERENCES providers(id),
    upstream_model VARCHAR(200) NOT NULL,
    input_price_per_1m NUMERIC(10, 4) NOT NULL DEFAULT 0,
    output_price_per_1m NUMERIC(10, 4) NOT NULL DEFAULT 0,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    max_tokens INT NOT NULL DEFAULT 4096,
    supports_stream BOOLEAN NOT NULL DEFAULT TRUE,
    supports_tools BOOLEAN NOT NULL DEFAULT FALSE,
    supports_vision BOOLEAN NOT NULL DEFAULT FALSE
);

CREATE TABLE IF NOT EXISTS usage_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    api_key_id BIGINT,
    model_name VARCHAR(100) NOT NULL,
    provider_name VARCHAR(50),
    prompt_tokens INT NOT NULL DEFAULT 0,
    completion_tokens INT NOT NULL DEFAULT 0,
    total_tokens INT NOT NULL DEFAULT 0,
    cost NUMERIC(12, 8) NOT NULL DEFAULT 0,
    latency NUMERIC(8, 3) NOT NULL DEFAULT 0,
    status VARCHAR(20) NOT NULL DEFAULT 'ok',
    http_status INT NOT NULL DEFAULT 200,
    is_stream BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_usage_logs_user_id ON usage_logs(user_id);
CREATE INDEX idx_usage_logs_model_name ON usage_logs(model_name);
CREATE INDEX idx_usage_logs_created_at ON usage_logs(created_at);

-- Seed default providers
INSERT INTO providers (name, base_url) VALUES
    ('openai', 'https://api.openai.com/v1'),
    ('anthropic', 'https://api.anthropic.com/v1'),
    ('google', 'https://generativelanguage.googleapis.com/v1beta'),
    ('deepseek', 'https://api.deepseek.com/v1'),
    ('alibaba', 'https://dashscope.aliyuncs.com/compatible-mode/v1'),
    ('bytedance', 'https://ark.cn-beijing.volces.com/api/v3')
ON CONFLICT (name) DO NOTHING;
