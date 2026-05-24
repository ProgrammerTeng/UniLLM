-- UniLLM: Migrate from Geneasy proxy to direct API keys
-- Run this when you have direct API keys from each provider
--
-- Steps:
--   1. Fill in your API keys below (replace 'YOUR_KEY_HERE')
--   2. Run: psql $DATABASE_URL < scripts/migrate_to_direct.sql
--   3. Restart the server to pick up new providers

BEGIN;

-- 1. Register direct providers
INSERT INTO providers (name, base_url, is_active, created_at) VALUES
  ('openai',    'https://api.openai.com/v1',                    true, NOW()),
  ('anthropic', 'https://api.anthropic.com/v1',                 true, NOW()),
  ('google',    'https://generativelanguage.googleapis.com/v1beta', true, NOW()),
  ('deepseek',  'https://api.deepseek.com/v1',                  true, NOW())
ON CONFLICT (name) DO UPDATE SET is_active = true, base_url = EXCLUDED.base_url;

-- 2. Add API keys (FILL IN YOUR KEYS)
-- OpenAI
INSERT INTO provider_keys (provider_id, key_value, weight, is_active, rpm, created_at)
SELECT id, 'YOUR_OPENAI_KEY_HERE', 1, true, 60, NOW() FROM providers WHERE name = 'openai';

-- Anthropic
INSERT INTO provider_keys (provider_id, key_value, weight, is_active, rpm, created_at)
SELECT id, 'YOUR_ANTHROPIC_KEY_HERE', 1, true, 60, NOW() FROM providers WHERE name = 'anthropic';

-- Google
INSERT INTO provider_keys (provider_id, key_value, weight, is_active, rpm, created_at)
SELECT id, 'YOUR_GOOGLE_KEY_HERE', 1, true, 60, NOW() FROM providers WHERE name = 'google';

-- DeepSeek
INSERT INTO provider_keys (provider_id, key_value, weight, is_active, rpm, created_at)
SELECT id, 'YOUR_DEEPSEEK_KEY_HERE', 1, true, 60, NOW() FROM providers WHERE name = 'deepseek';

-- 3. Re-point Claude models to Anthropic direct
UPDATE model_configs SET
  provider_id = (SELECT id FROM providers WHERE name = 'anthropic'),
  upstream_model = CASE public_name
    WHEN 'claude-opus-4-6'   THEN 'claude-opus-4-6'
    WHEN 'claude-sonnet-4-6' THEN 'claude-sonnet-4-6'
    WHEN 'claude-sonnet-4.5' THEN 'claude-sonnet-4-5-20250929'
    WHEN 'claude-haiku-4.5'  THEN 'claude-haiku-4-5-20251001'
    WHEN 'claude-3-7-sonnet' THEN 'claude-3-7-sonnet-20250219'
    ELSE upstream_model
  END
WHERE public_name LIKE 'claude%';

-- 4. Re-point Gemini models to Google direct
UPDATE model_configs SET
  provider_id = (SELECT id FROM providers WHERE name = 'google'),
  upstream_model = public_name  -- Google uses same names
WHERE public_name LIKE 'gemini%';

-- 5. Re-point DeepSeek to direct
UPDATE model_configs SET
  provider_id = (SELECT id FROM providers WHERE name = 'deepseek'),
  upstream_model = 'deepseek-chat'  -- direct API naming
WHERE public_name = 'deepseek-v3.2';

-- 6. Add OpenAI models (not available on Geneasy)
INSERT INTO model_configs (public_name, provider_id, upstream_model, input_price_per_1m, output_price_per_1m, is_active, max_tokens, supports_stream, supports_tools, supports_vision)
SELECT val.public_name, p.id, val.upstream_model, val.input_price, val.output_price, true, val.max_tok, true, val.tools, val.vision
FROM providers p,
(VALUES
  ('gpt-4o',      'gpt-4o',             2.5,  10.0, 128000, true,  true),
  ('gpt-4o-mini', 'gpt-4o-mini',        0.15,  0.6, 128000, true,  true),
  ('gpt-4.1',     'gpt-4.1',            2.0,   8.0, 128000, true,  false),
  ('o3',          'o3',                 10.0,  40.0, 200000, true,  false),
  ('o3-mini',     'o3-mini',             1.1,   4.4, 200000, true,  false),
  ('o4-mini',     'o4-mini',             1.1,   4.4, 200000, true,  false)
) AS val(public_name, upstream_model, input_price, output_price, max_tok, tools, vision)
WHERE p.name = 'openai'
ON CONFLICT (public_name) DO UPDATE SET
  provider_id = EXCLUDED.provider_id,
  upstream_model = EXCLUDED.upstream_model;

-- 7. Disable Geneasy provider (keep data for reference)
UPDATE providers SET is_active = false WHERE name = 'geneasy';

COMMIT;

-- Verify migration
SELECT mc.public_name, p.name AS provider, mc.upstream_model
FROM model_configs mc JOIN providers p ON mc.provider_id = p.id
WHERE mc.is_active = true
ORDER BY p.name, mc.public_name;
