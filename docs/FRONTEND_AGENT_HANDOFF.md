# Handoff: фронтенд CyberMate

Backend: `tgapp-` (Go). UI — **отдельный** репозиторий (Vite/React и т.д.).

## CORS и API URL

Локально — в `.env` бэкенда:

```env
CORS_ALLOWED_ORIGINS=http://localhost:5173
```

На фронте:

```env
VITE_API_URL=http://localhost:8090
# или VITE_API_BASE_URL — как в вашем репозитории фронта
```

**Продакшен (Railway):** см. `docs/PRODUCTION_RAILWAY.md` — `PORT`, Postgres, `CORS_ALLOWED_ORIGINS`, `VITE_API_BASE_URL` при сборке фронта.

## API: тема (light / dark)

| Метод | URL |
|-------|-----|
| GET | `/v1/users/telegram/{telegram_id}` → `data.theme` |
| PATCH | `/v1/users/telegram/{telegram_id}/theme` body: `{"theme":"dark"}` |

Константы: `light`, `dark`. localStorage: `cybermate-ui-theme`.

Эталон UI: `docs/frontend-theme-reference/` (`index.html`, `app.css`, `app.js`).

При старте: localStorage → API → `Telegram.WebApp.colorScheme` → `light`.

При смене: `applyTheme` + PATCH; в Mini App — `Telegram.WebApp.openTelegramLink` не нужен для темы.

## API: кнопка Support и реферальная ссылка

| Метод | URL |
|-------|-----|
| GET | `/v1/app/links` → `support_chat_url`, `bot_username`, `referral_link_base` |
| GET | `/v1/users/telegram/{telegram_id}/referral-link` → `referral_link` |
| GET | `/v1/users/telegram/{telegram_id}/referrals` → список приглашённых + `total_count`, `total_earnings` |

Бот Mini App: **@CyberMate_bot** (`TELEGRAM_BOT_USERNAME=CyberMate_bot`).

Реферальная ссылка (Mini App): `https://t.me/CyberMate_bot?startapp=ref_{telegram_id}` — не `CyberMateBot` и не `?start=`.

По умолчанию support: `https://t.me/+jXI2qDR9Y-xkZTI6` ([CyberMate Community](https://t.me/+jXI2qDR9Y-xkZTI6)).

```ts
export function openSupport() {
  const url = import.meta.env.VITE_SUPPORT_URL ?? "https://t.me/+jXI2qDR9Y-xkZTI6";
  window.Telegram?.WebApp?.openTelegramLink(url) ?? window.open(url, "_blank");
}
```

Подробнее: `docs/FRONTEND_SUPPORT_BUTTON.md`.

## Миграции и тестовый пользователь

```powershell
.\scripts\migrate.ps1
```

Telegram Web App test user: `telegram_id=777000` — нужен профиль в БД (`POST /v1/register` или seed).

## Чеклист фронт-агента

- [ ] `VITE_API_URL` + CORS origin в backend `.env`
- [ ] Тема: `data-theme`, CSS variables, PATCH theme API
- [ ] Support: `openTelegramLink` на `support_chat_url`
- [ ] Реферал: `GET .../referral-link` или `referral_link_base` + `telegram_id` (бот @CyberMate_bot)
- [ ] Список рефералов: `GET .../referrals` — см. `docs/FRONTEND_REFERRAL_PAGE.md`
- [ ] Не встраивать статику в Go (`pkg/web` не используется)
