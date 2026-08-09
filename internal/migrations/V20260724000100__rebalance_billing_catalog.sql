-- Rebalance the billing catalog (subscription plans + coin packs) to the values
-- shipped in pkg/billing/defaults.go on 2026-07-24.
--
-- Context: plans/packs live in admin_settings (JSONB) under keys
-- 'subscription_plans' and 'coin_packs'. When those rows are absent the service
-- falls back to the Go defaults automatically, so fresh installs need no change.
-- This migration only rewrites ALREADY-SEEDED rows to the new balanced values.
--
-- The migrate scripts re-run every migration on each invocation and there is no
-- Flyway-style history table, so this migration is guarded by a dedicated
-- data_migrations bookkeeping table and applies exactly once. That guard also
-- ensures admin edits made through the panel AFTER this migration ran are never
-- overwritten by a later `migrate` re-run.

CREATE TABLE IF NOT EXISTS data_migrations (
    name        TEXT PRIMARY KEY,
    applied_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM data_migrations
        WHERE name = 'V20260724000100__rebalance_billing_catalog'
    ) THEN
        RETURN;
    END IF;

    -- Subscription plans: only if a seeded row exists (fresh installs keep Go defaults).
    UPDATE admin_settings
    SET value = $plans$[
        {
            "id": "free", "name": "Старт", "badge": "Бесплатно", "badge_class": "free",
            "price_rub": 0, "price_sub": "навсегда", "coins": 15, "popular": false,
            "enabled": true, "sort_order": 1,
            "features": [
                "15 монет для старта",
                "YandexGPT, GPT OSS 20B, DeepSeek Chat",
                "FLUX (изображения)",
                "Qwen3 TTS, OmniVoice, MiniMax Speech"
            ],
            "locked": ["Видео — нет", "3D — нет", "Премиум модели — нет"]
        },
        {
            "id": "basic", "name": "Базовый", "badge": "Доступный", "badge_class": "basic",
            "price_rub": 149, "price_sub": "/ месяц", "coins": 160, "popular": false,
            "enabled": true, "sort_order": 2,
            "features": [
                "160 монет / месяц",
                "Claude Haiku, Gemini Flash, GPT-4o mini, DeepSeek Flash",
                "Nano Banana, Alice AI, Seedream, Qwen Image, Z-Image",
                "Kling Standard, Hailuo T2V (≈2–3 видео)",
                "ElevenLabs, Hunyuan 3D rapid"
            ],
            "locked": ["Pro/Max видео и 3D — нет"]
        },
        {
            "id": "pro", "name": "Про", "badge": "Популярный", "badge_class": "popular",
            "price_rub": 349, "price_sub": "/ месяц", "coins": 400, "popular": true,
            "enabled": true, "sort_order": 3,
            "features": [
                "400 монет / месяц",
                "Claude Sonnet, GPT-5.4, DeepSeek R1, Qwen 3.6",
                "GPT Image 2, Nano Banana 2, Grok Imagine",
                "Kling Pro, Seedance, WAN, Vidu, HappyHorse (≈4 видео)",
                "Mureka, ACE-Step, Tripo, Meshy 3D"
            ],
            "locked": []
        },
        {
            "id": "max", "name": "Максимум", "badge": "Выгодный", "badge_class": "max",
            "price_rub": 799, "price_sub": "/ месяц", "coins": 950, "popular": false,
            "enabled": true, "sort_order": 4,
            "features": [
                "950 монет / месяц",
                "GPT-4o, Gemini 2.5 Pro, Claude Opus 4.7, o3",
                "Nano Banana Pro",
                "Kling 4K, Seedance 2.0, Sora, Veo (≈6 видео)",
                "Tripo H3.1, Rodin 3D"
            ],
            "locked": []
        },
        {
            "id": "ultra", "name": "Бизнес", "badge": "Для бизнеса", "badge_class": "biz",
            "price_rub": 1999, "price_sub": "/ месяц", "coins": 2600, "popular": false,
            "enabled": true, "sort_order": 5,
            "features": [
                "2600 монет / месяц",
                "Claude Opus 4.8, o1, GPT-5.5, Sora Pro",
                "Все модели без ограничений",
                "Максимальный приоритет",
                "Лучшая цена за монету"
            ],
            "locked": []
        }
    ]$plans$::jsonb,
        updated_at = now()
    WHERE key = 'subscription_plans';

    -- Coin packs: pack-500 replaced by pack-300, pack-2500 added, prices raised to
    -- sit just above the subscription per-coin rate.
    UPDATE admin_settings
    SET value = $packs$[
        {"id": "pack-100",  "name": "100 монет",  "coins": 100,  "price_rub": 129,  "badge": "",     "enabled": true, "sort_order": 1},
        {"id": "pack-300",  "name": "300 монет",  "coins": 300,  "price_rub": 349,  "badge": "−10%", "enabled": true, "sort_order": 2},
        {"id": "pack-1000", "name": "1000 монет", "coins": 1000, "price_rub": 1049, "badge": "−18%", "enabled": true, "sort_order": 3},
        {"id": "pack-2500", "name": "2500 монет", "coins": 2500, "price_rub": 2399, "badge": "−26%", "enabled": true, "sort_order": 4}
    ]$packs$::jsonb,
        updated_at = now()
    WHERE key = 'coin_packs';

    INSERT INTO data_migrations(name)
    VALUES ('V20260724000100__rebalance_billing_catalog');
END $$;
