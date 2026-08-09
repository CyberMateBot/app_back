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
// Balance model (see docs/PRICING_REPORT.md):
//   - 1 CyberCoin = 1 ₽ base rate. Per-model prices are set at ~3.5× provider cost
//     (model_prices.go), so a spent coin costs us ≈ 1/3.5 ≈ 0.29 ₽ in provider fees.
//   - Coins are granted once per billing period (on purchase / admin grant) and are
//     spendable only on models the tier unlocks (gating.go).
//   - Coin allowance is sized so each tier can actually use its flagship feature
//     several times per month (the previous allowances were below the price of a
//     single unlocked video). Per-coin rate falls as the tier rises, so a higher
//     subscription is always the cheapest way to buy coins; one-time packs sit at a
//     small premium above subscriptions (see DefaultCoinPacks).
//
// Per-coin rate: basic 0.93 ₽, pro 0.87 ₽, max 0.84 ₽, ultra 0.77 ₽.
func DefaultSubscriptionPlans() []SubscriptionPlan {
	return []SubscriptionPlan{
		{
			ID: "free", Name: "Старт", Badge: "Бесплатно", BadgeClass: "free",
			PriceRub: 0, PriceSub: "навсегда", Coins: 15, SortOrder: 1, Enabled: true,
			Features: []string{"15 монет для старта", "YandexGPT, GPT OSS 20B, DeepSeek Chat", "FLUX (изображения)", "Qwen3 TTS, OmniVoice, MiniMax Speech"},
			Locked:   []string{"Видео — нет", "3D — нет", "Премиум модели — нет"},
		},
		{
			ID: "basic", Name: "Базовый", Badge: "Доступный", BadgeClass: "basic",
			PriceRub: 149, PriceSub: "/ месяц", Coins: 160, SortOrder: 2, Enabled: true,
			Features: []string{"160 монет / месяц", "Claude Haiku, Gemini Flash, GPT-4o mini, DeepSeek Flash", "Nano Banana, Alice AI, Seedream, Qwen Image, Z-Image", "Kling Standard, Hailuo T2V (≈2–3 видео)", "ElevenLabs, Hunyuan 3D rapid"},
			Locked:   []string{"Pro/Max видео и 3D — нет"},
		},
		{
			ID: "pro", Name: "Про", Badge: "Популярный", BadgeClass: "popular",
			PriceRub: 349, PriceSub: "/ месяц", Coins: 400, SortOrder: 3, Enabled: true, Popular: true,
			Features: []string{"400 монет / месяц", "Claude Sonnet, GPT-5.4, DeepSeek R1, Qwen 3.6", "GPT Image 2, Nano Banana 2, Grok Imagine", "Kling Pro, Seedance, WAN, Vidu, HappyHorse (≈4 видео)", "Mureka, ACE-Step, Tripo, Meshy 3D"},
		},
		{
			ID: "max", Name: "Максимум", Badge: "Выгодный", BadgeClass: "max",
			PriceRub: 799, PriceSub: "/ месяц", Coins: 950, SortOrder: 4, Enabled: true,
			Features: []string{"950 монет / месяц", "GPT-4o, Gemini 2.5 Pro, Claude Opus 4.7, o3", "Nano Banana Pro", "Kling 4K, Seedance 2.0, Sora, Veo (≈6 видео)", "Tripo H3.1, Rodin 3D"},
		},
		{
			ID: "ultra", Name: "Бизнес", Badge: "Для бизнеса", BadgeClass: "biz",
			PriceRub: 1999, PriceSub: "/ месяц", Coins: 2600, SortOrder: 5, Enabled: true,
			Features: []string{"2600 монет / месяц", "Claude Opus 4.8, o1, GPT-5.5, Sora Pro", "Все модели без ограничений", "Максимальный приоритет", "Лучшая цена за монету"},
		},
	}
}

// DefaultCoinPacks are seeded when admin has not configured packs yet.
//
// One-time top-ups sit at a small premium above the subscription per-coin rate, so
// subscribing is always the better deal, while larger packs approach the 1 ₽ base
// rate. Per-coin rate: 100→1.29 ₽, 300→1.16 ₽, 1000→1.05 ₽, 2500→0.96 ₽.
func DefaultCoinPacks() []CoinPack {
	return []CoinPack{
		{ID: "pack-100", Name: "100 монет", Coins: 100, PriceRub: 129, SortOrder: 1, Enabled: true},
		{ID: "pack-300", Name: "300 монет", Coins: 300, PriceRub: 349, Badge: "−10%", SortOrder: 2, Enabled: true},
		{ID: "pack-1000", Name: "1000 монет", Coins: 1000, PriceRub: 1049, Badge: "−18%", SortOrder: 3, Enabled: true},
		{ID: "pack-2500", Name: "2500 монет", Coins: 2500, PriceRub: 2399, Badge: "−26%", SortOrder: 4, Enabled: true},
	}
}
