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

var defaultPrices = map[string]int{
	"yandexgpt":        5,
	"gpt-oss-20b":      6,
	"gpt-oss-120b":     10,
	"qwen3.6-35b":      5,
	"qwen3-235b":       5,
	"claude-haiku-4.5": 10,
	"claude-sonnet-4.5": 10,
	"gemini-2.5-flash": 4,
	"gemini-2.5-pro":   8,
	"nano-banana":      35,
	"gpt-image-2":      20,
	"flux-dev":         8,
	"alice-ai-art":     15,
	"kling":            90,
	"seedance":         70,
	"qwen3-tts":        8,
	"mureka":           12,
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
