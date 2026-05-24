-- UniLLM Seed Data: Geneasy.AI as upstream provider
-- Usage: psql $DATABASE_URL < scripts/seed_geneasy.sql
--
-- This configures Geneasy (maas.geneasy.ai) as the sole upstream provider.
-- All models are routed through Geneasy's OpenAI-compatible endpoint.
-- Switch to direct API keys later by adding new providers and updating model_configs.

BEGIN;

-- 1. Provider: Geneasy (OpenAI-compatible protocol)
INSERT INTO providers (name, base_url, is_active, created_at)
VALUES ('geneasy', 'https://maas.geneasy.ai/v1', true, NOW())
ON CONFLICT (name) DO UPDATE SET base_url = EXCLUDED.base_url, is_active = true;

-- 2. Provider API Key
INSERT INTO provider_keys (provider_id, key_value, weight, is_active, rpm, created_at)
SELECT id, 'sk-REPLACE_WITH_GENEASY_KEY', 1, true, 60, NOW()
FROM providers WHERE name = 'geneasy'
ON CONFLICT DO NOTHING;

-- 3. Model Configs
-- Pricing: Geneasy markup price (USD per 1M tokens).
-- Adjust later when switching to direct API keys.

-- Claude models (Geneasy's naming convention)
INSERT INTO model_configs (public_name, provider_id, upstream_model, input_price_per_1m, output_price_per_1m, is_active, max_tokens, supports_stream, supports_tools, supports_vision)
SELECT val.public_name, p.id, val.upstream_model, val.input_price, val.output_price, true, val.max_tok, true, val.tools, val.vision
FROM providers p,
(VALUES
  -- Claude
  ('claude-opus-4-6',              'claude-opus-4-6',              15.0,  75.0,  200000, true,  false),
  ('claude-sonnet-4-6',            'claude-sonnet-4-6',             3.0,  15.0,  200000, true,  false),
  ('claude-sonnet-4.5',            'claude-4-5-sonnet-20250929',    3.0,  15.0,  200000, true,  false),
  ('claude-haiku-4.5',             'claude-4-5-haiku-20251001',     0.8,   4.0,  200000, true,  false),
  ('claude-3-7-sonnet',            'claude-3-7-sonnet-20250219',    3.0,  15.0,  200000, true,  false),
  -- Gemini
  ('gemini-2.5-flash',             'gemini-2.5-flash',              0.15,  0.60, 100000, true,  false),
  ('gemini-2.5-pro',               'gemini-2.5-pro',                1.25,  5.00, 100000, true,  false),
  ('gemini-2.5-flash-lite',        'gemini-2.5-flash-lite',         0.075, 0.30, 100000, true,  false),
  ('gemini-3-pro-preview',         'gemini-3-pro-preview',          1.25,  5.00, 100000, true,  false),
  -- DeepSeek
  ('deepseek-v3.2',                'deepseek/deepseek-v3.2',        0.27,  1.10, 128000, true,  false),
  -- Embedding (if Geneasy supports)
  ('text-embedding-3-small',       'text-embedding-3-small',        0.02,  0.0,  8191,   false, false),
  ('text-embedding-3-large',       'text-embedding-3-large',        0.13,  0.0,  8191,   false, false)
) AS val(public_name, upstream_model, input_price, output_price, max_tok, tools, vision)
WHERE p.name = 'geneasy'
ON CONFLICT (public_name) DO UPDATE SET
  upstream_model = EXCLUDED.upstream_model,
  input_price_per_1m = EXCLUDED.input_price_per_1m,
  output_price_per_1m = EXCLUDED.output_price_per_1m,
  is_active = true;

-- 4. Create admin user (change password after first login!)
INSERT INTO users (email, password_hash, name, role, balance, created_at, updated_at)
VALUES ('admin@unillm.com', '$2a$10$placeholder_change_me', 'Admin', 'admin', 1000.0, NOW(), NOW())
ON CONFLICT (email) DO NOTHING;

COMMIT;

-- Verify
SELECT '=== Providers ===' AS info;
SELECT id, name, base_url, is_active FROM providers;

SELECT '=== Provider Keys ===' AS info;
SELECT pk.id, p.name AS provider, pk.is_active, pk.rpm
FROM provider_keys pk JOIN providers p ON pk.provider_id = p.id;

SELECT '=== Models ===' AS info;
SELECT mc.public_name, p.name AS provider, mc.upstream_model,
       mc.input_price_per_1m AS in_price, mc.output_price_per_1m AS out_price,
       mc.supports_stream, mc.supports_tools
FROM model_configs mc JOIN providers p ON mc.provider_id = p.id
WHERE mc.is_active = true
ORDER BY p.name, mc.public_name;
