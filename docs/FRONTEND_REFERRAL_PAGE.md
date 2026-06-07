# Инструкция: страница реферальной программы (фронтенд)

Backend: репозиторий `tgapp-` (Go), порт по умолчанию `8090`.  
UI — **отдельный** репозиторий (Vite/React и т.д.).

## Цель

На странице реферальной программы показывать и копировать **правильную** ссылку приглашения в Mini App бота **@CyberMate_bot**, а не старую ссылку на другого бота.

---

## Что было неправильно (удалить с фронта)

| Ошибка | Пример |
|--------|--------|
| Старый username бота | `CyberMateBot` |
| Параметр `start` вместо `startapp` | `?start=777000` |
| Сборка ссылки вручную без API | `` `https://t.me/CyberMateBot?startapp=ref_${id}` `` |
| Support-ссылка вместо реферальной | `https://t.me/+jXI2qDR9Y-xkZTI6` |

**Правильный формат:**

```
https://t.me/CyberMate_bot?startapp=ref_{telegram_id}
```

Пример для тестового пользователя: `https://t.me/CyberMate_bot?startapp=ref_777000`

---

## API бэкенда

Базовый URL: `import.meta.env.VITE_API_URL` (например `http://localhost:8090`).

### 1. Готовая реферальная ссылка (предпочтительно)

```http
GET /v1/users/telegram/{telegram_id}/referral-link
```

**Ответ:**

```json
{
  "referral_link": "https://t.me/CyberMate_bot?startapp=ref_123456789"
}
```

`{telegram_id}` — ID текущего пользователя из `Telegram.WebApp.initDataUnsafe.user.id` (строка или число, в URL передать как строку).

### 2. Конфиг приложения (кэш на старте)

```http
GET /v1/app/links
```

**Ответ:**

```json
{
  "support_chat_url": "https://t.me/+jXI2qDR9Y-xkZTI6",
  "bot_username": "CyberMate_bot",
  "referral_link_base": "https://t.me/CyberMate_bot?startapp=ref_"
}
```

### 3. Список приглашённых рефералов

```http
GET /v1/users/telegram/{telegram_id}/referrals
```

`{telegram_id}` — ID **текущего** пользователя (пригласившего), не реферала.

**Ответ:**

```json
{
  "referrals": [
    {
      "profile_id": 42,
      "telegram_id": "987654321",
      "name": "Ivan",
      "username": "ivan",
      "avatar": "https://...",
      "completed_tasks_count": 2,
      "earnings": 150
    }
  ],
  "total_count": 1,
  "total_earnings": 150,
  "completed_tasks": 2
}
```

| Поле | Описание |
|------|----------|
| `referrals` | Массив приглашённых пользователей |
| `total_count` | Количество рефералов |
| `total_earnings` | Сумма `earnings` по всем рефералам |
| `completed_tasks` | Сумма `completed_tasks_count` |

Если рефералов нет — `referrals: []`, счётчики `0`.  
`404` — профиль с таким `telegram_id` не найден.

---

## Рекомендуемая реализация

### Файл `src/lib/referral.ts` (или аналог)

```ts
const API = import.meta.env.VITE_API_URL ?? "http://localhost:8090";

export function getTelegramUserId(): string | null {
  const id = window.Telegram?.WebApp?.initDataUnsafe?.user?.id;
  return id != null ? String(id) : null;
}

export async function fetchReferralLink(telegramId: string): Promise<string> {
  const res = await fetch(
    `${API}/v1/users/telegram/${encodeURIComponent(telegramId)}/referral-link`
  );
  if (!res.ok) {
    throw new Error(`referral-link failed: ${res.status}`);
  }
  const data = (await res.json()) as { referral_link?: string };
  if (!data.referral_link) {
    throw new Error("referral_link is empty");
  }
  return data.referral_link;
}

export async function copyReferralLink(link: string): Promise<void> {
  const tg = window.Telegram?.WebApp;
  if (navigator.clipboard?.writeText) {
    await navigator.clipboard.writeText(link);
    tg?.showAlert?.("Ссылка скопирована");
    return;
  }
  tg?.showAlert?.(link);
}

export type ReferralItem = {
  profile_id: number;
  telegram_id: string;
  name: string;
  username: string;
  avatar: string;
  completed_tasks_count: number;
  earnings: number;
};

export type ReferralsResponse = {
  referrals: ReferralItem[];
  total_count: number;
  total_earnings: number;
  completed_tasks: number;
};

export async function fetchReferrals(telegramId: string): Promise<ReferralsResponse> {
  const res = await fetch(
    `${API}/v1/users/telegram/${encodeURIComponent(telegramId)}/referrals`
  );
  if (!res.ok) {
    throw new Error(`referrals failed: ${res.status}`);
  }
  return (await res.json()) as ReferralsResponse;
}
```

### Страница реферальной программы

1. При монтировании / открытии экрана:
   - взять `telegramId` из `getTelegramUserId()`;
   - если нет ID — показать ошибку («Откройте через Telegram»);
   - параллельно вызвать `fetchReferralLink(telegramId)` и `fetchReferrals(telegramId)`;
   - сохранить в state (`referralLink`, `referrals`, `total_count`, `total_earnings`).

2. Поле «Ваша ссылка» / QR — показывать **только** `referralLink` с API, не хардкод.

3. Блок «Мои рефералы» — список из `referrals`:
   - если `total_count === 0` — пустое состояние («Пока никого не пригласили»);
   - для каждого элемента: имя (`name` или `@username`), `completed_tasks_count`, `earnings`;
   - аватар — `avatar`, если пустой — fallback по первой букве имени.

4. Кнопка «Скопировать» → `copyReferralLink(referralLink)`.

5. Кнопка «Поделиться» (опционально):

```ts
const tg = window.Telegram?.WebApp;
const url = referralLink;
const text = "Присоединяйся к CyberMate";
if (tg?.openTelegramLink) {
  tg.openTelegramLink(
    `https://t.me/share/url?url=${encodeURIComponent(url)}&text=${encodeURIComponent(text)}`
  );
} else if (navigator.share) {
  await navigator.share({ title: "CyberMate", text, url });
}
```

5. **Не** использовать `<a href={referralLink} target="_blank">` внутри Mini App для открытия своей же реферальной ссылки — для внешних t.me-ссылок в Telegram клиенте предпочтительнее `Telegram.WebApp.openTelegramLink(url)`.

Пример загрузки данных на странице:

```tsx
useEffect(() => {
  const telegramId = getTelegramUserId();
  if (!telegramId) return;

  Promise.all([
    fetchReferralLink(telegramId),
    fetchReferrals(telegramId),
  ]).then(([link, stats]) => {
    setReferralLink(link);
    setReferrals(stats.referrals);
    setTotalCount(stats.total_count);
    setTotalEarnings(stats.total_earnings);
  }).catch(console.error);
}, []);
```

---

## Регистрация по реферальной ссылке (бэкенд)

Когда новый пользователь открывает Mini App по ссылке `?startapp=ref_777000`, в `initData` приходит `start_param` (часто `ref_777000`).

При регистрации фронт передаёт его в API (рекомендуется явно):

```http
POST /v1/register
Content-Type: application/json

{
  "init_data_raw": "<base64 initData>",
  "start_param": "ref_777000"
}
```

Бэкенд:
- извлекает `777000` из префикса `ref_` и создаёт запись в `referrals`;
- если `start_param` не передан в body, читает его из `init_data_raw`;
- если пользователь уже зарегистрирован (`409 AlreadyExists`), но пришёл по реферальной ссылке — **всё равно** привязывает реферала (идемпотентно).

На странице реферальной программы убедиться, что при `POST /v1/register` не обнуляется `start_param` из `Telegram.WebApp.initDataUnsafe.start_param`.

---

## Поиск и правки в коде фронта

Найти и **заменить** все вхождения:

```bash
# примеры grep
CyberMateBot
t.me/CyberMateBot
startapp=ref_
?start=
referral
invite
shareLink
```

Типичные файлы: `Referral*.tsx`, `Invite*.tsx`, `Profile*.tsx`, константы `BOT_USERNAME`, `REFERRAL_URL`.

---

## Проверка после правок

1. `VITE_API_URL` указывает на запущенный бэкенд; в `.env` бэкенда `CORS_ALLOWED_ORIGINS` включает origin фронта (например `http://localhost:5173`).

2. В DevTools → Network на странице рефералов:
   - запрос `GET .../referral-link` → `200`;
   - запрос `GET .../referrals` → `200` (или `404`, если профиль ещё не создан);
   - в ответе `referral_link` содержит `CyberMate_bot` и `startapp=ref_`.

3. Скопированная ссылка открывается в Telegram и ведёт в **ваш** Mini App (@CyberMate_bot), не в @CyberMateBot.

4. Тест: открыть ссылку с другого аккаунта → регистрация → `GET .../referrals` у пригласившего возвращает нового пользователя в `referrals`.

---

## Чеклист для фронт-агента

- [ ] Удалён хардкод `CyberMateBot` и ручная сборка `t.me/...`
- [ ] Ссылка загружается с `GET /v1/users/telegram/{id}/referral-link` (или `referral_link_base` + id)
- [ ] Список рефералов загружается с `GET /v1/users/telegram/{id}/referrals`
- [ ] В UI отображается полная ссылка с `startapp=ref_`
- [ ] Блок «Мои рефералы» показывает `referrals` из API (не мок/локальный массив)
- [ ] «Копировать» / «Поделиться» используют значение с API
- [ ] `POST /v1/register` передаёт `start_param` из Telegram initData
- [ ] Support (`t.me/+...`) не смешан с реферальной ссылкой на бота

---

## Связанные документы

- `docs/FRONTEND_AGENT_HANDOFF.md` — общий handoff (CORS, тема, support)
- `docs/FRONTEND_SUPPORT_BUTTON.md` — кнопка Support (отдельно от реферала)
