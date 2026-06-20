package admincatalog

import "github.com/twelvepills-936/tgapp-/pkg/ai"

// Entry is a model row for the admin panel.
type Entry struct {
	ID       string
	Name     string
	Provider string
	Category string
	Price    int
	Enabled  bool
}

// defaultPrices is the cost in CyberCoins per operation.
// 1 CyberCoin = 1 ₽. Prices calculated at 3.5× provider cost to ensure
// 65-75% margin. Provider rates: WaveSpeed ($/M tokens × 90 ₽/$),
// Yandex (₽/1000 tokens). Typical request: 1500 in + 700 out tokens.
var defaultPrices = map[string]int{
	// === TEXT · Yandex provider ===
	"yandexgpt":   5, // cost ~1.76 ₽/req → 5 coins (3× margin)
	"gpt-oss-20b": 1, // cost ~0.22 ₽/req
	"gpt-oss-120b": 3, // cost ~0.66 ₽/req
	"qwen3.6-35b":  4, // cost ~1.1 ₽/req
	"qwen3-235b":   4, // cost ~1.1 ₽/req

	// === TEXT · WaveSpeed provider ===
	"deepseek-v4-flash":     1,  // cost ~0.05 ₽/req
	"deepseek-chat":         1,  // cost ~0.06 ₽/req
	"deepseek-v3.2":         1,  // cost ~0.07 ₽/req
	"deepseek-chat-v3-0324": 1,  // cost ~0.07 ₽/req
	"deepseek-v3.2-exp":     2,  // cost ~0.09 ₽/req
	"deepseek-v4":           2,  // cost ~0.36 ₽/req
	"deepseek-r1":           3,  // cost ~0.45 ₽/req
	"gpt-4.1-nano":          1,  // cost ~0.04 ₽/req
	"gpt-4o-mini":           1,  // cost ~0.07 ₽/req
	"gpt-5.4-mini":          2,  // cost ~0.36 ₽/req
	"gpt-4.1-mini":          2,  // cost ~0.15 ₽/req
	"o4-mini":               4,  // cost ~0.54 ₽/req
	"gemini-2.5-flash":      2,  // cost ~0.20 ₽/req
	"claude-haiku-4.5":      2,  // cost ~0.45 ₽/req
	"gpt-4.1":               4,  // cost ~0.70 ₽/req
	"gpt-4o":                5,  // cost ~0.92 ₽/req
	"gemini-2.5-pro":        4,  // cost ~0.81 ₽/req
	"o3-mini":               5,  // cost ~1.08 ₽/req
	"gpt-5.4":               5,  // cost ~1.28 ₽/req
	"claude-sonnet-4.5":     5,  // cost ~1.28 ₽/req
	"o3":                   12,  // cost ~2.70 ₽/req
	"o1":                   15,  // cost ~3.15 ₽/req
	"claude-opus-4.7":       8,  // cost ~2.16 ₽/req
	"claude-opus-4.8":       8,  // cost ~2.16 ₽/req
	"gpt-5.5":              10,  // cost ~2.56 ₽/req

	// === IMAGE · WaveSpeed ===
	"flux-dev":         5,  // $0.015 → 1.35 ₽
	"nano-banana":     10,  // $0.035 → 3.15 ₽
	"gpt-image-2":     20,  // $0.057 → 5.1 ₽
	"gpt-image-1.5":   20,  // $0.057 → 5.1 ₽
	"nano-banana-2":   22,  // $0.063 → 5.7 ₽
	"nano-banana-pro": 40,  // $0.126 → 11.3 ₽

	// === IMAGE · Yandex ===
	"alice-ai-art": 12, // ~3.5 ₽/image

	// === VIDEO · WaveSpeed (per ~5s clip) ===
	"kling":              80,  // Kling V3 Std $0.25 → 22.5 ₽
	"kling-v3-std":       80,
	"kling-v3-pro":      110,  // $0.35 → 31.5 ₽
	"kling-v3-4k":       160,  // $0.50 → 45 ₽
	"seedance":           95,  // Seedance V1 $0.30 → 27 ₽
	"seedance-v1-pro-i2v": 95,
	"seedance-v1.5-i2v-fast": 110, // $0.35 → 31.5 ₽
	"seedance-v1.5-t2v-fast": 110,
	"seedance-v1.5-i2v-spicy": 110,
	"seedance-v2-video-edit":   160, // $0.51 → 45.9 ₽
	"seedance-v2-video-extend": 160,

	// === AUDIO · WaveSpeed ===
	"qwen3-tts":          4,  // ~$0.01 → 0.9 ₽
	"omnivoice":          3,  // ~$0.008 → 0.72 ₽
	"elevenlabs-v3":      8,  // ~$0.025 → 2.25 ₽
	"minimax-speech-2.6": 3,  // ~$0.008 → 0.72 ₽
	"mureka":            50,  // song ~$0.20 → 18 ₽
	"mureka-v9":         50,
	"ace-step-1.5":      40,  // track ~$0.15 → 13.5 ₽

	// === 3D · WaveSpeed ===
	"hunyuan3d-v3.1-rapid": 25, // ~$0.08 → 7.2 ₽
	"hunyuan3d-v3-t2d":     30, // ~$0.10 → 9 ₽
	"tripo3d-v2.5-i2d":     48, // ~$0.15 → 13.5 ₽
	"tripo3d-v2.5-multiview": 48,
	"tripo3d-h3.1-t2d":     55, // ~$0.18 → 16 ₽
	"tripo3d-h3.1-i2d":     55,
	"meshy6-t2d":           48, // ~$0.15 → 13.5 ₽
	"rodin-v2-i2d":         55, // ~$0.18 → 16 ₽
	"rodin-v2.5-i2d":       55,
}

func providerLabel(provider string) string {
	switch provider {
	case "yandex":
		return "Yandex"
	case "wavespeed":
		return "WaveSpeed"
	default:
		if provider == "" {
			return "—"
		}
		return provider
	}
}

func defaultPrice(id, category string) int {
	if price, ok := defaultPrices[id]; ok {
		return price
	}
	switch category {
	case "text":
		return 5
	case "video":
		return 70
	case "audio":
		return 8
	case "3d":
		return 50
	default:
		return 15
	}
}

// BaseEntries returns the catalog merged from AI packages with default prices.
func BaseEntries() []Entry {
	out := make([]Entry, 0, 64)

	for _, model := range ai.ListTextModels() {
		out = append(out, Entry{
			ID:       model.ID,
			Name:     model.Label,
			Provider: providerLabel(model.Provider),
			Category: "text",
			Price:    defaultPrice(model.ID, "text"),
			Enabled:  true,
		})
	}

	appendMedia := func(models []ai.MediaModel) {
		for _, model := range models {
			out = append(out, Entry{
				ID:       model.ID,
				Name:     model.Label,
				Provider: providerLabel(model.Provider),
				Category: model.Kind,
				Price:    defaultPrice(model.ID, model.Kind),
				Enabled:  true,
			})
		}
	}

	appendMedia(ai.ListImageModels())
	appendMedia(ai.ListVideoModels())
	appendMedia(ai.ListAudioModels())
	appendMedia(ai.ListThreeDModels())

	return out
}
