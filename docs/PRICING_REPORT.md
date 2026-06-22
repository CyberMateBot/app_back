# CyberMate Pricing Report

Дата актуализации: 2026-06-22

Документ описывает тарифы подписок, начисление монет (CyberCoins), матрицу доступа к AI-моделям и техническую реализацию ограничений.

## Принципы

1. **Подписка = доступ к моделям.** Монеты списываются за каждую генерацию отдельно.
2. **1 CyberCoin = 1 ₽** (базовый курс; паки монет могут давать скидку).
3. **Срок подписки:** по умолчанию 30 дней; админ может выдать любой срок или бессрочно.
4. **Истечение:** по `expires_at` пользователь автоматически переходит на `free`; премиум-модели блокируются на backend и в UI.
5. **Источник истины для gating:** `app_back/pkg/billing/gating.go` (frontend зеркало: `app_front/src/lib/planGating.js`).

---

## Тарифы подписок

| План | ID | Цена | Монет / период | Ранг |
|------|-----|------|----------------|------|
| Старт | `free` | 0 ₽ | 10 | 0 |
| Базовый | `basic` | 149 ₽/мес | 40 | 1 |
| Про | `pro` | 349 ₽/мес | 100 | 2 |
| Максимум | `max` | 799 ₽/мес | 250 | 3 |
| Бизнес | `ultra` | 1999 ₽/мес | 600 | 4 |

> Монеты при покупке подписки начисляются один раз за период (или при ручной выдаче админом с флагом `grant_coins`). Значения снижены относительно ранних версий, чтобы лучше соответствовать реальной стоимости генераций.

### Что разблокирует каждый план

| Категория | Free | Basic | Pro | Max | Ultra |
|-----------|------|-------|-----|-----|-------|
| Text | 3 дешёвые модели | + fast / mini модели | + GPT-5.4, Sonnet | + GPT-4o, o3, Opus 4.7 | + Opus 4.8, o1, GPT-5.5 |
| Image | FLUX | + Alice, Nano Banana | + GPT Image 2 | + Nano Banana Pro | все |
| Video | — | Kling Standard | Kling Pro, Seedance | Kling 4K, Seedance 2 | все |
| Audio | базовые TTS | ElevenLabs | Mureka, ACE-Step | все | все |
| 3D | — | Hunyuan rapid | Tripo, Meshy | Tripo H3.1, Rodin | все |

---

## Пакеты монет (разово)

| ID | Монеты | Цена |
|----|--------|------|
| pack-100 | 100 | 99 ₽ |
| pack-500 | 500 | 449 ₽ |
| pack-1000 | 1000 | 799 ₽ |

---

## Матрица доступа к моделям (min plan rank)

Ранг плана должен быть **≥** min rank модели.

### Text

| Min rank | План | Модели |
|----------|------|--------|
| 0 | free | yandexgpt, gpt-oss-20b, deepseek-chat |
| 1 | basic | gpt-4o-mini, gpt-oss-120b, qwen3-235b, deepseek-v4-flash, gpt-4.1-nano, claude-haiku-4.5, gemini-2.5-flash, o4-mini, … |
| 2 | pro | gpt-4.1, gpt-5.4, claude-sonnet-4.5, deepseek-r1, o3-mini, … |
| 3 | max | gpt-4o, gemini-2.5-pro, claude-opus-4.7, o3 |
| 4 | ultra | claude-opus-4.8, o1, gpt-5.5 |

### Image

| Min rank | План | Модели |
|----------|------|--------|
| 0 | free | flux-dev |
| 1 | basic | alice-ai-art, nano-banana |
| 2 | pro | gpt-image-1.5, gpt-image-2, nano-banana-2 |
| 3 | max | nano-banana-pro |

### Video (free недоступен)

| Min rank | План | Модели |
|----------|------|--------|
| 1 | basic | kling-v3-std |
| 2 | pro | kling-v3-pro, seedance-v1-* |
| 3 | max | kling-v3-4k, seedance-v2-* |

### Audio

| Min rank | План | Модели |
|----------|------|--------|
| 0 | free | omnivoice, minimax-speech-2.6, qwen3-tts |
| 1 | basic | elevenlabs-v3 |
| 2 | pro | mureka, mureka-v9, ace-step-1.5 |

### 3D (free недоступен)

| Min rank | План | Модели |
|----------|------|--------|
| 1 | basic | hunyuan3d-v3.1-rapid, hunyuan3d-v3-t2d |
| 2 | pro | tripo3d-v2.5-*, meshy6-t2d |
| 3 | max | tripo3d-h3.1-*, rodin-v2-* |

**Fallback для неизвестных model ID:** text=free, image/video/3d=basic, audio=free.

---

## Примеры цен генерации (CyberCoins)

Источник: `app_back/pkg/billing/model_prices.go`

| Модель | Категория | Монеты |
|--------|-----------|--------|
| gpt-oss-20b | text | 1 |
| yandexgpt | text | 5 |
| gpt-4o | text | 5 |
| claude-opus-4.8 | text | 8 |
| flux-dev | image | 5 |
| nano-banana | image | 10 |
| kling-v3-std | video | 80 |
| kling-v3-pro | video | 110 |
| kling-v3-4k | video | 160 |
| qwen3-tts | audio | 4 |
| hunyuan3d-v3.1-rapid | 3d | 25 |

---

## Техническая реализация

### Backend

| Компонент | Путь |
|-----------|------|
| Таблица подписок | `user_subscriptions` (migration `V20260622000100__user_subscriptions.sql`) |
| Логика срока / expiry | `internal/usecase/subscription.go` |
| Gating при генерации | `pkg/tokenguard/guard.go` → `CheckAccessForModel` |
| API состояния подписки | `GET /v1/users/telegram/{id}/subscription` |
| Каталог тарифов | `GET /v1/billing/catalog` |
| Админ: выдача подписки | `POST /api/admin/users/{id}/subscription` |
| Админ: сброс | `DELETE /api/admin/users/{id}/subscription` |

Поля выдачи подписки админом:
- `plan_id` — free | basic | pro | max | ultra
- `duration_days` — срок от текущего момента
- `expires_at` — точная дата (RFC3339), приоритет над duration_days
- `no_expiry` — бессрочно
- `grant_coins` — начислить монеты плана

### Frontend (Mini App)

| Функция | Реализация |
|---------|------------|
| Загрузка подписки | `fetchUserSubscription` → state `subscriptionState` |
| Блокировка моделей в UI | `planGating.js`, `AiVariantSelect`, каталог |
| Уведомления | колокольчик на главной → dropdown (`AppNotifications`) |
| Виджет срока | `SubscriptionTimeBadge` на главной и «Подписки» |
| Expiring soon | `expiring_soon=true` если ≤ 3 дней (`ExpiringSoonDays`) |

### Admin Console

| Функция | UI |
|---------|-----|
| Выдача / редактирование подписки | Users → «Подписка» (`UserSubscriptionModal`) |
| Редактирование тарифов и монет | Pricing page |

---

## Проверка (QA)

1. Пользователь **free** не может сгенерировать видео / 3D / nano-banana (403 на backend, lock в UI).
2. Пользователь **basic** может Kling Standard, но не Kling Pro.
3. После истечения `expires_at` API возвращает `plan_id=free`, модели Pro блокируются.
4. За 3 дня до конца — `expiring_soon=true`, уведомление в колокольчике.
5. Админ выдал подписку на 7 дней — в приложении отображается «Осталось 7 дн.»

---

## Файлы для синхронизации при изменении цен

1. `app_back/pkg/billing/defaults.go` — монеты и описания планов
2. `app_back/pkg/billing/gating.go` — доступ к моделям
3. `app_back/pkg/billing/model_prices.go` — цены генераций
4. `app_front/src/lib/billing.js` — fallback каталога
5. `app_front/src/lib/planGating.js` — зеркало gating для UI
6. `app_front/src/lib/modelPrices.js` — зеркало цен для UI
