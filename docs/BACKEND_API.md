# Admin API — спецификация для бэкенд-агента

Базовый префикс: `/api/admin`

Авторизация: заголовок `Authorization: Bearer <admin_jwt>` на всех эндпоинтах, кроме логина.

Формат ошибок (единый):

```json
{
  "error": "human readable message",
  "code": "OPTIONAL_ERROR_CODE"
}
```

---

## Auth

### `POST /api/admin/auth/login`

**Body:**
```json
{
  "email": "admin@example.com",
  "password": "secret"
}
```

**Response 200:**
```json
{
  "token": "jwt-token",
  "admin": {
    "id": 1,
    "email": "admin@example.com"
  }
}
```

**Errors:** `401` — неверные credentials

---

### `POST /api/admin/auth/logout`

**Response:** `204 No Content`

---

### `GET /api/admin/auth/me`

**Response 200:**
```json
{
  "id": 1,
  "email": "admin@example.com"
}
```

---

## Stats

### `GET /api/admin/stats`

**Response 200:**
```json
{
  "total_users": 1000,
  "active_users_today": 42,
  "new_users_today": 5,
  "total_messages": 15000
}
```

---

## Users

### `GET /api/admin/users`

Список пользователей с пагинацией.

**Query params:**
| Param | Type | Default | Description |
|-------|------|---------|-------------|
| `page` | int | 1 | Номер страницы |
| `per_page` | int | 20 | Записей на страницу (max 100) |
| `search` | string | — | Поиск по `username`, `first_name`, `last_name` |

**Response 200:**
```json
{
  "data": [
    {
      "id": 1,
      "telegram_id": 123456789,
      "username": "ivan",
      "first_name": "Иван",
      "last_name": "Петров",
      "is_active": true,
      "tokens": 1500,
      "created_at": "2025-01-15T10:00:00Z"
    }
  ],
  "total": 100
}
```

**Требования к полю `tokens`:**
- `integer >= 0`
- обязательно в списке и в деталях пользователя

---

### `GET /api/admin/users/:id`

**Response 200:**
```json
{
  "id": 1,
  "telegram_id": 123456789,
  "username": "ivan",
  "first_name": "Иван",
  "last_name": "Петров",
  "is_active": true,
  "tokens": 1500,
  "created_at": "2025-01-15T10:00:00Z"
}
```

**Errors:** `404` — пользователь не найден

---

### `PATCH /api/admin/users/:id`

Блокировка / разблокировка.

**Body:**
```json
{
  "is_active": false
}
```

**Response 200:** объект пользователя (как в `GET /api/admin/users/:id`)

---

### `DELETE /api/admin/users/:id`

**Response:** `204 No Content`

**Errors:** `404` — пользователь не найден

---

## Tokens

Операции с балансом токенов. Баланс хранится в `wallets.balance_available` (профиль = user в admin API).

### `POST /api/admin/users/:id/tokens/credit`

**Body:**
```json
{
  "amount": 100,
  "reason": "Бонус за активность"
}
```

**Response 200:**
```json
{
  "user_id": 1,
  "tokens": 1600,
  "delta": 100,
  "operation": "credit"
}
```

**Errors:** `404`, `422` — невалидный `amount`

---

### `POST /api/admin/users/:id/tokens/debit`

**Body:**
```json
{
  "amount": 50,
  "reason": "Корректировка баланса"
}
```

**Response 200:**
```json
{
  "user_id": 1,
  "tokens": 1550,
  "delta": -50,
  "operation": "debit"
}
```

**Errors:** `404`, `422` — недостаточно токенов или невалидный `amount`

---

## Broadcast

### `POST /api/admin/broadcast`

**Body:**
```json
{
  "message": "Текст рассылки",
  "target": "all",
  "parse_mode": "HTML"
}
```

`target`: `"all"` | `"active"`

**Response 200:**
```json
{
  "sent": 950,
  "failed": 2
}
```

---

## Dev

Бэкенд на `http://127.0.0.1:8090`, Vite proxy `/api` → бэкенд.
