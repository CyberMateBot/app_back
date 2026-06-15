# AI models — фронтенд

## Список моделей

```http
GET /v1/generate/models
```

```json
{
  "text_models": [
    {
      "id": "yandexgpt",
      "label": "YandexGPT",
      "group": "Yandex",
      "description": "Быстрая универсальная модель...",
      "tier": "standard",
      "provider": "yandex"
    },
    {
      "id": "gpt-oss-20b",
      "label": "GPT OSS 20B",
      "group": "Open-weight GPT",
      "tier": "fast",
      "provider": "yandex"
    }
  ]
}
```

`tier`: `fast` | `standard` | `pro` — для бейджей «Быстро» / «Баланс» / «Макс. качество».

## Генерация текста

```http
POST /v1/generate/text
Content-Type: application/json

{
  "prompt": "Напиши пост про CyberMate",
  "model": "gpt-oss-120b",
  "temperature": 0.3,
  "max_tokens": 2000
}
```

`model` — поле **`id`** из `GET /v1/generate/models` (не путать с OpenAI как брендом).

## Рекомендуемый UI

| Группа в приложении | Модели | Как показать |
|---------------------|--------|--------------|
| **Yandex** | YandexGPT | Основной выбор для RU |
| **DeepSeek** | V4, R1, V4 Flash, V3.2, V3.2 Exp, V3 0324, Chat | Код и рассуждения через Wavespeed |
| **ChatGPT** | GPT-5.4, 5.4 Mini, 4.1, 4.1 Mini/Nano, 4o, 4o Mini, o4-mini, o3, o3-mini, o1 | OpenAI через Wavespeed |
| **Claude** | Haiku, Sonnet, Opus | Anthropic через Wavespeed |
| **Gemini** | 2.5 Flash, 2.5 Pro | Google через Wavespeed |
| **Open-weight GPT** | GPT OSS 20B (fast), GPT OSS 120B (pro) | Один пункт «GPT OSS» → подменю Fast / Pro |
| **Qwen** | Qwen3.6 35B (standard), Qwen3 235B (pro) | Один пункт «Qwen» → Баланс / Максимум |

**Не называйте** GPT OSS «OpenAI» — это вводит в заблуждение (OpenAI — другая компания). Правильно: **GPT OSS 20B** / **GPT OSS 120B**.

## Пример (React)

```ts
const API = import.meta.env.VITE_API_BASE_URL;

type TextModel = {
  id: string;
  label: string;
  group: string;
  tier: string;
};

export async function loadModels(): Promise<TextModel[]> {
  const r = await fetch(`${API}/v1/generate/models`);
  const data = await r.json();
  return data.text_models;
}

export async function generateText(prompt: string, modelId: string) {
  const r = await fetch(`${API}/v1/generate/text`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ prompt, model: modelId }),
  });
  if (!r.ok) throw new Error(await r.text());
  return r.json() as Promise<{ text: string; model: string }>;
}
```

Сохраняйте выбранный `modelId` в `localStorage`, чтобы не сбрасывать при перезаходе.

## Отображение ответа (Markdown + формулы)

Ответ `POST /v1/generate/text` содержит `"format": "markdown"`. **Нельзя** выводить `text` как plain text в `<pre>` или с классом ошибки (красный цвет).

### Рекомендуемый стек

```bash
npm install react-markdown remark-math rehype-katex katex
```

```tsx
import ReactMarkdown from "react-markdown";
import remarkMath from "remark-math";
import rehypeKatex from "rehype-katex";
import "katex/dist/katex.min.css";

export function AiMessage({ text }: { text: string }) {
  return (
    <div className="ai-message text-neutral-900 dark:text-neutral-100">
      <ReactMarkdown remarkPlugins={[remarkMath]} rehypePlugins={[rehypeKatex]}>
        {text}
      </ReactMarkdown>
    </div>
  );
}
```

### Почему был «красный текст» и `∗∗`

| Симптом | Причина |
|--------|---------|
| `D_f=\mathbb{R}...` без формул | Нет рендера LaTeX — нужен KaTeX |
| `∗∗Упрощение∗∗` | Unicode-звёздочки вместо `**` или Markdown не парсится |
| Всё красное | CSS класс ошибки на блок ответа (`.error`, `color: red`) |

Бэкенд нормализует `∗` → `*` и оборачивает «голый» LaTeX в `$...$`, но **рендер на фронте обязателен**.

### Стили

```css
.ai-message {
  color: inherit; /* не red */
  line-height: 1.5;
  word-break: break-word;
}
.ai-message .katex {
  font-size: 1em;
}
```

## Railway

Достаточно уже настроенных:

- `YANDEX_GPT_API_KEY`
- `YANDEX_GPT_FOLDER_ID`

Доп. переменные для новых моделей не нужны — slug зашиты в бэкенде.
