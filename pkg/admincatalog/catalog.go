package admincatalog

import (
	"github.com/twelvepills-936/tgapp-/pkg/ai"
	"github.com/twelvepills-936/tgapp-/pkg/billing"
)

// Entry is a model row for the admin panel.
type Entry struct {
	ID       string
	Name     string
	Provider string
	Category string
	Price    int
	Enabled  bool
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

// DefaultPrice returns the CyberCoin cost for a model operation.
func DefaultPrice(id, category string) int {
	return billing.DefaultModelPrice(id, category)
}

func defaultPrice(id, category string) int {
	return billing.DefaultModelPrice(id, category)
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
