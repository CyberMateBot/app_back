package ai

import (
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/billing"
)

// TextModel describes a selectable text generation model for the frontend.
type TextModel struct {
	ID            string `json:"id"`
	Label         string `json:"label"`
	Group         string `json:"group"`
	Description   string `json:"description,omitempty"`
	Tier          string `json:"tier"` // fast | standard | pro
	Provider      string `json:"provider"` // yandex | wavespeed
	SupportsImage bool   `json:"supports_image,omitempty"`
	MinPlan       string `json:"min_plan,omitempty"` // minimum subscription plan to use this model
}

type textModelDef struct {
	ID            string
	Label         string
	Group         string
	Description   string
	Tier          string
	Slug          string // provider model id / Yandex slug
	UseResponses  bool
	UseOpenAIChat bool
	UseWavespeed  bool
	SupportsImage bool
}

// textModelCatalog — Yandex AI Studio models.
var textModelCatalog = []textModelDef{
	{
		ID: "yandexgpt", Label: "YandexGPT", Group: "Yandex",
		Description: "Повседневные вопросы, письма и тексты на русском",
		Tier: "standard", Slug: "yandexgpt/latest",
	},
	{
		ID: "gpt-oss-20b", Label: "GPT OSS 20B", Group: "Open-weight GPT",
		Description: "Черновики, короткие ответы и быстрые правки текста",
		Tier: "fast", Slug: "gpt-oss-20b/latest", UseResponses: true,
	},
	{
		ID: "gpt-oss-120b", Label: "GPT OSS 120B", Group: "Open-weight GPT",
		Description: "Сложные задачи, развёрнутые ответы и глубокие рассуждения",
		Tier: "pro", Slug: "gpt-oss-120b/latest", UseResponses: true,
	},
	{
		ID: "qwen3.6-35b", Label: "Qwen3.6 35B", Group: "Qwen",
		Description: "Точные ответы, структура и работа с длинным контекстом",
		Tier: "pro", Slug: "qwen3.6-35b-a3b", UseOpenAIChat: true, SupportsImage: true,
	},
	{
		ID: "qwen3-235b", Label: "Qwen3 235B", Group: "Qwen",
		Description: "Быстрые сводки и черновики по большим объёмам текста",
		Tier: "fast", Slug: "qwen3-235b-a22b-fp8/latest", UseResponses: true,
	},
}

// wavespeedTextModelCatalog — Wavespeed LLM (OpenAI-compatible chat/completions).
var wavespeedTextModelCatalog = []textModelDef{
	{
		ID: "claude-haiku-4.5", Label: "Claude Haiku 4.5", Group: "Claude",
		Description: "Быстрые ответы и повседневные задачи",
		Tier: "fast", Slug: "anthropic/claude-haiku-4.5", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "claude-sonnet-4.5", Label: "Claude Sonnet 4.5", Group: "Claude",
		Description: "Баланс скорости и качества для сложных задач",
		Tier: "standard", Slug: "anthropic/claude-sonnet-4.5", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "claude-opus-4.7", Label: "Claude Opus 4.7", Group: "Claude",
		Description: "Максимальное качество рассуждений и анализа",
		Tier: "pro", Slug: "anthropic/claude-opus-4.7", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "claude-opus-4.8", Label: "Claude Opus 4.8", Group: "Claude",
		Description: "Топовая модель Claude для сложнейших задач",
		Tier: "pro", Slug: "anthropic/claude-opus-4.8", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "gemini-2.5-flash", Label: "Gemini 2.5 Flash", Group: "Gemini",
		Description: "Быстрые мультимодальные ответы Google AI",
		Tier: "fast", Slug: "google/gemini-2.5-flash", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "gemini-2.5-pro", Label: "Gemini 2.5 Pro", Group: "Gemini",
		Description: "Глубокий анализ, код и работа с изображениями",
		Tier: "pro", Slug: "google/gemini-2.5-pro", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "gpt-5.4", Label: "GPT-5.4", Group: "ChatGPT",
		Description: "Самая новая модель OpenAI для сложных задач",
		Tier: "pro", Slug: "openai/gpt-5.4", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "gpt-5.4-mini", Label: "GPT-5.4 Mini", Group: "ChatGPT",
		Description: "Быстрые и экономичные ответы GPT-5.4",
		Tier: "fast", Slug: "openai/gpt-5.4-mini", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "gpt-4.1", Label: "GPT-4.1", Group: "ChatGPT",
		Description: "Длинный контекст и глубокий анализ документов",
		Tier: "pro", Slug: "openai/gpt-4.1", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "gpt-4.1-mini", Label: "GPT-4.1 Mini", Group: "ChatGPT",
		Description: "Баланс цены и качества для повседневных задач",
		Tier: "standard", Slug: "openai/gpt-4.1-mini", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "gpt-4.1-nano", Label: "GPT-4.1 Nano", Group: "ChatGPT",
		Description: "Максимальная скорость для коротких ответов",
		Tier: "fast", Slug: "openai/gpt-4.1-nano", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "gpt-4o", Label: "GPT-4o", Group: "ChatGPT",
		Description: "Омни-модель: текст, изображения и инструменты",
		Tier: "pro", Slug: "openai/gpt-4o", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "gpt-4o-mini", Label: "GPT-4o Mini", Group: "ChatGPT",
		Description: "Популярная быстрая модель для чата и черновиков",
		Tier: "fast", Slug: "openai/gpt-4o-mini", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "o4-mini", Label: "o4-mini", Group: "ChatGPT",
		Description: "O-серия: низкая задержка и пошаговые рассуждения",
		Tier: "fast", Slug: "openai/o4-mini", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "o3", Label: "o3", Group: "ChatGPT",
		Description: "Мощные рассуждения для сложных задач",
		Tier: "pro", Slug: "openai/o3", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "o3-mini", Label: "o3-mini", Group: "ChatGPT",
		Description: "O-серия, оптимизирована для программирования",
		Tier: "pro", Slug: "openai/o3-mini", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "o1", Label: "o1", Group: "ChatGPT",
		Description: "Оригинальная O-серия с chain-of-thought",
		Tier: "pro", Slug: "openai/o1", UseWavespeed: true, SupportsImage: true,
	},
	{
		ID: "deepseek-v4", Label: "DeepSeek V4", Group: "DeepSeek",
		Description: "Лучшая модель для кодинга и сложных задач",
		Tier: "pro", Slug: "deepseek/deepseek-v4", UseWavespeed: true,
	},
	{
		ID: "deepseek-r1", Label: "DeepSeek R1", Group: "DeepSeek",
		Description: "Пошаговые рассуждения для сложных задач и логики",
		Tier: "pro", Slug: "deepseek/deepseek-r1", UseWavespeed: true,
	},
	{
		ID: "deepseek-v4-flash", Label: "DeepSeek V4 Flash", Group: "DeepSeek",
		Description: "Быстрые и экономичные ответы DeepSeek V4",
		Tier: "fast", Slug: "deepseek/deepseek-v4-flash", UseWavespeed: true,
	},
	{
		ID: "deepseek-v3.2", Label: "DeepSeek V3.2", Group: "DeepSeek",
		Description: "Сбалансированная модель для текста и кода",
		Tier: "standard", Slug: "deepseek/deepseek-v3.2", UseWavespeed: true,
	},
	{
		ID: "deepseek-v3.2-exp", Label: "DeepSeek V3.2 Exp", Group: "DeepSeek",
		Description: "Экспериментальная версия DeepSeek V3.2",
		Tier: "pro", Slug: "deepseek/deepseek-v3.2-exp", UseWavespeed: true,
	},
	{
		ID: "deepseek-chat-v3-0324", Label: "DeepSeek V3 0324", Group: "DeepSeek",
		Description: "Улучшенная версия DeepSeek V3",
		Tier: "standard", Slug: "deepseek/deepseek-chat-v3-0324", UseWavespeed: true,
	},
	{
		ID: "deepseek-chat", Label: "DeepSeek Chat", Group: "DeepSeek",
		Description: "Базовая модель DeepSeek для повседневных задач",
		Tier: "fast", Slug: "deepseek/deepseek-chat", UseWavespeed: true,
	},
}

// modelAliases maps client model ids to catalog ids.
var modelAliases = map[string]string{
	"yandex": "yandexgpt", "default": "yandexgpt",
	"gpt-oss-120b/latest": "gpt-oss-120b", "gpt_oss_120b": "gpt-oss-120b",
	"gpt-oss-20b/latest": "gpt-oss-20b", "gpt_oss_20b": "gpt-oss-20b",
	"qwen3.6-35b-a3b/latest": "qwen3.6-35b", "qwen3.6-35b": "qwen3.6-35b",
	"qwen3-235b-a22b-fp8/latest": "qwen3-235b", "qwen3-235b": "qwen3-235b",
	"deepseek": "deepseek-v3.2",
	"deepseek-v32/latest": "deepseek-v3.2", "deepseek-v32": "deepseek-v3.2",
	"deepseek/deepseek-v4": "deepseek-v4",
	"deepseek/deepseek-r1": "deepseek-r1",
	"deepseek/deepseek-v4-flash": "deepseek-v4-flash",
	"deepseek/deepseek-v3.2": "deepseek-v3.2",
	"deepseek/deepseek-v3.2-exp": "deepseek-v3.2-exp",
	"deepseek/deepseek-chat-v3-0324": "deepseek-chat-v3-0324",
	"deepseek/deepseek-chat": "deepseek-chat",
	// legacy / shorthand
	"gemini": "gemini-2.5-flash", "gemini-flash": "gemini-2.5-flash",
	"anthropic/claude-haiku-4.5": "claude-haiku-4.5",
	"anthropic/claude-sonnet-4.5": "claude-sonnet-4.5",
	"anthropic/claude-opus-4.7": "claude-opus-4.7",
	"anthropic/claude-opus-4.8": "claude-opus-4.8",
	"google/gemini-2.5-flash": "gemini-2.5-flash",
	"google/gemini-2.5-pro": "gemini-2.5-pro",
	"openai": "gpt-4o-mini",
	"openai/gpt-5.4": "gpt-5.4",
	"openai/gpt-5.4-mini": "gpt-5.4-mini",
	"openai/gpt-4.1": "gpt-4.1",
	"openai/gpt-4.1-mini": "gpt-4.1-mini",
	"openai/gpt-4.1-nano": "gpt-4.1-nano",
	"openai/gpt-4o": "gpt-4o",
	"openai/gpt-4o-mini": "gpt-4o-mini",
	"openai/o4-mini": "o4-mini",
	"openai/o3": "o3",
	"openai/o3-mini": "o3-mini",
	"openai/o1": "o1",
}

func ListTextModels() []TextModel {
	out := make([]TextModel, 0, len(textModelCatalog)+len(wavespeedTextModelCatalog))
	for _, m := range textModelCatalog {
		out = append(out, toTextModel(m, "yandex"))
	}
	for _, m := range wavespeedTextModelCatalog {
		out = append(out, toTextModel(m, "wavespeed"))
	}
	return out
}

func toTextModel(m textModelDef, provider string) TextModel {
	return TextModel{
		ID: m.ID, Label: m.Label, Group: m.Group,
		Description: m.Description, Tier: m.Tier, Provider: provider,
		SupportsImage: m.SupportsImage,
		MinPlan:       billing.MinPlanForModel(m.ID, "text"),
	}
}

func resolveTextModel(requested string, cfgSlug string) (def textModelDef, ok bool) {
	key := normalizeModelKey(requested)
	if key == "" {
		key = "yandexgpt"
	}
	if id, found := modelAliases[key]; found {
		key = id
	}
	key = strings.TrimPrefix(key, "gpt://")
	if i := strings.Index(key, "/"); i > 0 {
		key = key[strings.LastIndex(key, "/")+1:]
	}
	key = strings.TrimSuffix(key, "/latest")

	if def, ok = findTextModelDef(wavespeedTextModelCatalog, key); ok {
		return def, true
	}
	if def, ok = findTextModelDef(textModelCatalog, key); ok {
		return def, true
	}
	// legacy: env default slug (yandex only)
	if cfgSlug != "" && (key == cfgSlug || strings.Contains(cfgSlug, key)) {
		return textModelDef{ID: key, Label: key, Slug: cfgSlug, Tier: "standard"}, true
	}
	return textModelDef{}, false
}

func findTextModelDef(catalog []textModelDef, key string) (textModelDef, bool) {
	for _, m := range catalog {
		if m.ID == key || strings.TrimSuffix(m.Slug, "/latest") == key || m.Slug == key {
			return m, true
		}
	}
	return textModelDef{}, false
}

func normalizeModelKey(requested string) string {
	return strings.ToLower(strings.TrimSpace(requested))
}
