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
func DefaultSubscriptionPlans() []SubscriptionPlan {
	return []SubscriptionPlan{
		{
			ID: "free", Name: "Старт", Badge: "Бесплатно", BadgeClass: "free",
			PriceRub: 0, PriceSub: "навсегда", Coins: 50, SortOrder: 1, Enabled: true,
			Features: []string{"50 монет / месяц", "Базовые чат-модели", "10 изображений (FLUX)", "TTS (базовый)"},
			Locked:   []string{"Видео и музыка — нет", "Премиум модели — нет"},
		},
		{
			ID: "basic", Name: "Базовый", Badge: "Доступный", BadgeClass: "basic",
			PriceRub: 149, PriceSub: "/ месяц", Coins: 250, SortOrder: 2, Enabled: true,
			Features: []string{"250 монет / месяц", "Все fast-модели", "25 изображений HD", "3 видео (Kling Std)", "Музыка и озвучка"},
		},
		{
			ID: "pro", Name: "Про", Badge: "Популярный", BadgeClass: "popular",
			PriceRub: 349, PriceSub: "/ месяц", Coins: 700, SortOrder: 3, Enabled: true, Popular: true,
			Features: []string{"700 монет / месяц", "Claude, Gemini, GPT-5.4", "35 изображений (GPT Image 2)", "8 видео (Kling/Seedance)", "Музыка + 3D", "Приоритетная очередь"},
		},
		{
			ID: "max", Name: "Максимум", Badge: "Выгодный", BadgeClass: "max",
			PriceRub: 799, PriceSub: "/ месяц", Coins: 2000, SortOrder: 4, Enabled: true,
			Features: []string{"2000 монет / месяц", "Все PRO модели", "100 изображений (4K)", "25 видео HD", "Все инструменты", "Перенос 20% монет"},
		},
		{
			ID: "ultra", Name: "Бизнес", Badge: "Для бизнеса", BadgeClass: "biz",
			PriceRub: 1999, PriceSub: "/ месяц", Coins: 6000, SortOrder: 5, Enabled: true,
			Features: []string{"6000 монет / месяц", "Claude Opus, o3 и всё", "300+ изображений", "75 видео 4K", "API доступ", "Перенос 50% монет"},
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
