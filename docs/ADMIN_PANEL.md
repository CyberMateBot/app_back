# Admin Panel — подключение к бэкенду

Backend: репозиторий `tgapp-` (Go), порт `8090`.  
Admin Panel: отдельный репозиторий (`AdminPanel`, Vite/React).

## API

Базовый префикс: **`/api/admin`**

| Метод | Путь | Auth |
|-------|------|------|
| `POST` | `/api/admin/auth/login` | нет |
| `POST` | `/api/admin/auth/logout` | Bearer |
| `GET` | `/api/admin/auth/me` | Bearer |
| `GET` | `/api/admin/stats` | Bearer |
| `GET` | `/api/admin/users?page=1&per_page=20&search=` | Bearer |
| `GET` | `/api/admin/users/:id` | Bearer |
| `PATCH` | `/api/admin/users/:id` body `{"is_active": false}` | Bearer |
| `DELETE` | `/api/admin/users/:id` | Bearer |
| `POST` | `/api/admin/broadcast` | Bearer |

Форматы ответов совпадают с ожиданиями Admin Panel.

## Локальная разработка

### 1. Бэкенд `.env`

```env
APP_HTTP_PORT=8090
JWT_SECRET=your-secret-key

# Первый админ создаётся автоматически при старте, если таблица admins пуста:
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=your-secure-password

# Для рассылки:
TELEGRAM_BOT_ENABLED=true
TELEGRAM_BOT_TOKEN=...
```

### 2. Миграции

```powershell
docker compose up -d
.\scripts\migrate.ps1
```

### 3. Запуск бэкенда

```powershell
go run cmd/service/main.go
```

### 4. Admin Panel `.env.development`

```env
VITE_API_BASE_URL=
```

Vite проксирует `/api` → `http://127.0.0.1:8090`.

### 5. Проверка

```powershell
curl -X POST http://127.0.0.1:8090/api/admin/auth/login `
  -H "Content-Type: application/json" `
  -d '{"email":"admin@example.com","password":"your-secure-password"}'
```

## Продакшен (Railway)

### Бэкенд (`app_back`)

```env
CORS_ALLOWED_ORIGINS=https://adminconsole-production-a33a.up.railway.app
ADMIN_EMAIL=admin@example.com
ADMIN_PASSWORD=...
JWT_SECRET=...
```

Миграции в Postgres (Railway Query), включая `V20260607000100__profile_is_active.sql`.

### Admin Panel (`adminconsole`)

**Обязательно** переменная **Runtime** (не только Build):

```env
API_BASE_URL=https://appback-production-6c0e.up.railway.app
```

Без неё фронт шлёт запросы на `/api` **своего** домена → серый экран / пустой dashboard.

Порт `8080` в логах Admin Panel — **нормально** (Railway `PORT`). Бэкенд слушает свой `PORT` отдельно.

Альтернатива: `VITE_API_BASE_URL` при **сборке** (Build Variables) — тогда нужен redeploy после смены URL.

## Статистика

| Поле | Источник |
|------|----------|
| `total_users` | `COUNT(profiles)` |
| `active_users_today` | уникальные `profile_id` в `prompt_history` за сегодня |
| `new_users_today` | `profiles` зарегистрированные сегодня |
| `total_messages` | `COUNT(prompt_history)` |

## Рассылка

Требует `TELEGRAM_BOT_ENABLED=true` и валидный `TELEGRAM_BOT_TOKEN`.

- `target: "all"` — все активные (`is_active = true`) пользователи
- `target: "active"` — активные с запросами в `prompt_history` за последние 7 дней

## Пользователи

- `first_name` ← `profiles.name`
- `last_name` — пустая строка (в схеме нет отдельного поля)
- `telegram_id` — число из `profiles.telegram_id`
- `is_active` — колонка `profiles.is_active` (миграция `V20260607000100`)
