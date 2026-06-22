package billing

// SubscriptionPlan is a monthly subscription tier shown in the app.
type SubscriptionPlan struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Badge      string   `json:"badge"`
	BadgeClass string   `json:"badge_class"`
	PriceRub   int64    `json:"price_rub"`
	PriceSub   string   `json:"price_sub"`
	Coins      int64    `json:"coins"`
	Features   []string `json:"features"`
	Locked     []string `json:"locked,omitempty"`
	Popular    bool     `json:"popular"`
	Enabled    bool     `json:"enabled"`
	SortOrder  int32    `json:"sort_order"`
}

// CoinPack is a one-time coin purchase pack.
type CoinPack struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Coins     int64  `json:"coins"`
	PriceRub  int64  `json:"price_rub"`
	Badge     string `json:"badge,omitempty"`
	Enabled   bool   `json:"enabled"`
	SortOrder int32  `json:"sort_order"`
}

// DefaultSubscriptionPlans are seeded when admin has not configured plans yet.
//
// Coins are granted once per billing period (on purchase / admin grant) and were
// rebalanced down to better match per-model prices. Each tier also unlocks a wider
// set of AI models (see gating.go): free gets only the cheapest text/image/audio
// models, paid tiers progressively unlock video, 3D and premium models.
func DefaultSubscriptionPlans() []SubscriptionPlan {
	return []SubscriptionPlan{
		{
			ID: "free", Name: "Старт", Badge: "Бесплатно", BadgeClass: "free",
			PriceRub: 0, PriceSub: "навсегда", Coins: 10, SortOrder: 1, Enabled: true,
			Features: []string{"10 монет / месяц", "YandexGPT, GPT OSS 20B, DeepSeek Chat", "FLUX (изображения)", "Qwen3 TTS, OmniVoice, MiniMax Speech"},
			Locked:   []string{"Видео — нет", "3D — нет", "Премиум модели — нет"},
		},
		{
			ID: "basic", Name: "Базовый", Badge: "Доступный", BadgeClass: "basic",
			PriceRub: 149, PriceSub: "/ месяц", Coins: 40, SortOrder: 2, Enabled: true,
			Features: []string{"40 монет / месяц", "Claude Haiku, Gemini Flash, GPT-4o mini, DeepSeek Flash", "Nano Banana, Alice AI, Seedream, Qwen Image, Z-Image", "Kling Standard, Hailuo T2V", "ElevenLabs, Hunyuan 3D rapid"},
			Locked:   []string{"Pro/Max видео и 3D — нет"},
		},
		{
			ID: "pro", Name: "Про", Badge: "Популярный", BadgeClass: "popular",
			PriceRub: 349, PriceSub: "/ месяц", Coins: 100, SortOrder: 3, Enabled: true, Popular: true,
			Features: []string{"100 монет / месяц", "Claude Sonnet, GPT-5.4, DeepSeek R1, Qwen 3.6", "GPT Image 2, Nano Banana 2, Grok Imagine", "Kling Pro, Seedance, WAN, Vidu, HappyHorse", "Mureka, ACE-Step, Tripo, Meshy 3D"},
		},
		{
			ID: "max", Name: "Максимум", Badge: "Выгодный", BadgeClass: "max",
			PriceRub: 799, PriceSub: "/ месяц", Coins: 250, SortOrder: 4, Enabled: true,
			Features: []string{"250 монет / месяц", "GPT-4o, Gemini 2.5 Pro, Claude Opus 4.7, o3", "Nano Banana Pro", "Kling 4K, Seedance 2.0, Sora, Veo", "Tripo H3.1, Rodin 3D"},
		},
		{
			ID: "ultra", Name: "Бизнес", Badge: "Для бизнеса", BadgeClass: "biz",
			PriceRub: 1999, PriceSub: "/ месяц", Coins: 600, SortOrder: 5, Enabled: true,
			Features: []string{"600 монет / месяц", "Claude Opus 4.8, o1, GPT-5.5, Sora Pro", "Все модели без ограничений", "Максимальный приоритет", "Все видео, 3D и аудио"},
		},
	}
}

// DefaultCoinPacks are seeded when admin has not configured packs yet.
func DefaultCoinPacks() []CoinPack {
	return []CoinPack{
		{ID: "pack-100", Name: "100 монет", Coins: 100, PriceRub: 99, SortOrder: 1, Enabled: true},
		{ID: "pack-500", Name: "500 монет", Coins: 500, PriceRub: 449, Badge: "−10%", SortOrder: 2, Enabled: true},
		{ID: "pack-1000", Name: "1000 монет", Coins: 1000, PriceRub: 799, Badge: "−20%", SortOrder: 3, Enabled: true},
	}
}
