# Продакшен: Railway (бэкенд) + фронт (Mini App)

## Симптом

В Mini App: **«Не удалось связаться с API (Load failed)»**.

Чаще всего:

1. Бэкенд на Railway **не слушает порт `PORT`** (сервис недоступен снаружи).
2. **Postgres** не подключён — процесс падает при старте.
3. **CORS** — origin фронта не в `CORS_ALLOWED_ORIGINS`.
4. На фронте при сборке не задан **`VITE_API_BASE_URL`** (или `VITE_API_URL`).

---

## Railway: переменные бэкенда

### Обязательно

| Переменная | Пример | Комментарий |
|------------|--------|-------------|
| `PORT` | *(задаёт Railway автоматически)* | HTTP-сервер слушает этот порт |
| `PG_HOST` | из Railway Postgres | |
| `PG_PORT` | `5432` | |
| `PG_USER` | `postgres` | |
| `PG_PASS` | *** | |
| `PG_DBNAME` | `railway` | |
| `CORS_ALLOWED_ORIGINS` | см. ниже | |

`APP_HTTP_PORT` можно не задавать — при наличии `PORT` используется он.

### CORS

Для отладки:

```env
CORS_ALLOWED_ORIGINS=*
```

Для продакшена (рекомендуется) — **точный URL фронта без слэша в конце**, через запятую:

```env
CORS_ALLOWED_ORIGINS=https://ваш-фронт.vercel.app,https://web.telegram.org
```

Подставьте реальный домен, где открывается Mini App (Vercel / Netlify / GitHub Pages).

### AI (генерация текста и изображений)

```env
YANDEX_GPT_API_KEY=...
YANDEX_GPT_FOLDER_ID=...
YANDEX_ALICE_AI_ART_MODEL=aliceai-image-art-3.0
# Nano Banana — отдельный провайдер (без ключа модель не работает):
WAVESPEED_API_KEY=...
```

DeepSeek и Claude/Gemini идут через Wavespeed (`WAVESPEED_API_KEY`, `GEMINI_API_BASE_URL`). YandexGPT и Alice AI ART — через Yandex (`YANDEX_GPT_*`).

### Опционально

```env
TELEGRAM_BOT_USERNAME=CyberMate_bot
TELEGRAM_BOT_ENABLED=true
TELEGRAM_BOT_POLLING=false
SWAGGER_ENABLED=false
```

`TELEGRAM_BOT_POLLING=false` — по умолчанию. Для `/start` задайте webhook на этот бэкенд:

`TELEGRAM_WEBHOOK_URL=https://YOUR-BACKEND.up.railway.app/v1/telegram/webhook`

При старте сервис вызовет `setWebhook`. Сообщение `/start`: «⚡️ CyberMate…» + кнопка «Открыть CyberMate».

Локально: `TELEGRAM_BOT_POLLING=true` (webhook снимается при старте).

Если `/start` всё ещё со старым текстом — в BotFather или на другом сервере остался старый webhook; укажите `TELEGRAM_WEBHOOK_URL` на текущий бэкенд.

---

## Проверка бэкенда после деплоя

```bash
curl https://appback-production-6c0e.up.railway.app/health
# ok

curl https://appback-production-6c0e.up.railway.app/v1/app/links
# JSON с bot_username, referral_link_base
```

Если **таймаут** — смотрите логи Railway: часто `failed to init postgres` или сервис слушал не тот порт (исправлено в коде: `PORT` → HTTP).

---

## Фронтенд (сборка)

При **build** на хостинге фронта:

```env
VITE_API_BASE_URL=https://appback-production-6c0e.up.railway.app
```

Если в коде используется другое имя — проверьте репозиторий фронта:

```bash
grep -r "VITE_API" src/
```

Должен совпадать с переменной в CI/CD. URL **без** завершающего `/`.

Пересоберите и задеплойте фронт после смены переменной.

---

## Чеклист

- [ ] Railway: Postgres plugin подключён, `PG_*` в Variables сервиса API
- [ ] Railway: Deploy успешен, `curl .../health` → `ok`
- [ ] Railway: `CORS_ALLOWED_ORIGINS=*` или URL фронта
- [ ] Фронт: `VITE_API_BASE_URL` = URL Railway API при сборке
- [ ] В Telegram открывается **новая** сборка фронта (кэш WebView)

---

## Миграции Postgres на Railway

Миграции **не применяются автоматически** при деплое. Выполните SQL в **Railway → Postgres → Query** (база `railway`).

### Таблица `prompt_history` (история промптов)

Симптом в логах:

```text
relation "prompt_history" does not exist (SQLSTATE 42P01)
```

Скопируйте и выполните (нужна уже существующая таблица `profiles`):

```sql
CREATE TABLE IF NOT EXISTS prompt_history (
    id BIGSERIAL PRIMARY KEY,
    profile_id BIGINT NOT NULL REFERENCES profiles(id) ON DELETE CASCADE,
    prompt TEXT NOT NULL,
    category TEXT NOT NULL DEFAULT 'general',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    model TEXT NOT NULL DEFAULT '',
    response TEXT NOT NULL DEFAULT '',
    session_id TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_prompt_history_profile_created_at
    ON prompt_history (profile_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_prompt_history_profile_session
    ON prompt_history (profile_id, session_id, created_at DESC);
```

Проверка:

```sql
SELECT COUNT(*) FROM prompt_history;
```

Файл миграции в репозитории: `internal/migrations/V20260527000000__prompt_history.sql`.

---

## Логи Railway

**Deployments → сервис API → View logs**

| Лог | Действие |
|-----|----------|
| `failed to init postgres` | Настроить `PG_HOST`, `PG_PASS`, `PG_DBNAME` |
| `failed to serve HTTP` | Проверить `PORT` / перезапуск |
| `relation "prompt_history" does not exist` | Выполнить SQL выше в Railway Postgres Query |
| Сервис стартует и сразу exit | Часто БД; смотрите строку выше |
