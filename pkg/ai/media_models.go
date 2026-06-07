package ai

import "strings"

// MediaModel describes a selectable image or video generation model.
type MediaModel struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Group       string `json:"group"`
	Description string `json:"description,omitempty"`
	Provider    string `json:"provider"` // wavespeed | yandex
	Kind        string `json:"kind"`     // image | video
}

type mediaModelDef struct {
	ID          string
	Label       string
	Group       string
	Description string
	Slug        string
	Provider    string
	Kind        string
}

var wavespeedImageModelCatalog = []mediaModelDef{
	{
		ID: "nano-banana", Label: "Nano Banana Pro", Group: "Google",
		Description: "Быстрые иллюстрации и картинки по тексту",
		Slug: "google/nano-banana-pro/text-to-image", Provider: "wavespeed", Kind: "image",
	},
	{
		ID: "flux-dev", Label: "FLUX Dev", Group: "WaveSpeed",
		Description: "Качественные изображения по текстовому описанию",
		Slug: "wavespeed-ai/flux-dev", Provider: "wavespeed", Kind: "image",
	},
}

var wavespeedVideoModelCatalog = []mediaModelDef{
	{
		ID: "kling-v3-std", Label: "Kling 3.0 Standard", Group: "Kling",
		Description: "Быстрая генерация видео по тексту",
		Slug: "kwaivgi/kling-v3.0-std/text-to-video", Provider: "wavespeed", Kind: "video",
	},
	{
		ID: "kling-v3-pro", Label: "Kling 3.0 Pro", Group: "Kling",
		Description: "Высокое качество видео по тексту",
		Slug: "kwaivgi/kling-v3.0-pro/text-to-video", Provider: "wavespeed", Kind: "video",
	},
}

var mediaModelAliases = map[string]string{
	"nano-banana": "nano-banana", "banana": "nano-banana", "wavespeed": "nano-banana",
	"google/nano-banana-pro": "nano-banana",
	"google/nano-banana-pro/text-to-image": "nano-banana",
	"flux-dev": "flux-dev", "wavespeed-ai/flux-dev": "flux-dev",
	"kling-v3-std": "kling-v3-std", "kling-v3.0-std": "kling-v3-std",
	"kwaivgi/kling-v3.0-std/text-to-video": "kling-v3-std",
	"kling-v3-pro": "kling-v3-pro", "kling-v3.0-pro": "kling-v3-pro",
	"kwaivgi/kling-v3.0-pro/text-to-video": "kling-v3-pro",
}

func ListImageModels() []MediaModel {
	out := make([]MediaModel, 0, len(wavespeedImageModelCatalog)+1)
	for _, m := range wavespeedImageModelCatalog {
		out = append(out, toMediaModel(m))
	}
	out = append(out, MediaModel{
		ID: "alice-ai-art", Label: "Alice AI ART", Group: "Yandex",
		Description: "Художественные изображения через Yandex ART",
		Provider: "yandex", Kind: "image",
	})
	return out
}

func ListVideoModels() []MediaModel {
	out := make([]MediaModel, 0, len(wavespeedVideoModelCatalog))
	for _, m := range wavespeedVideoModelCatalog {
		out = append(out, toMediaModel(m))
	}
	return out
}

func toMediaModel(m mediaModelDef) MediaModel {
	return MediaModel{
		ID: m.ID, Label: m.Label, Group: m.Group,
		Description: m.Description, Provider: m.Provider, Kind: m.Kind,
	}
}

func resolveWavespeedImageModel(requested string) (mediaModelDef, bool) {
	return resolveMediaModel(wavespeedImageModelCatalog, requested)
}

func resolveWavespeedVideoModel(requested string) (mediaModelDef, bool) {
	return resolveMediaModel(wavespeedVideoModelCatalog, requested)
}

func resolveMediaModel(catalog []mediaModelDef, requested string) (mediaModelDef, bool) {
	key := strings.ToLower(strings.TrimSpace(requested))
	if key == "" {
		return mediaModelDef{}, false
	}
	if id, ok := mediaModelAliases[key]; ok {
		key = id
	}
	for _, m := range catalog {
		if m.ID == key || strings.EqualFold(m.Slug, key) {
			return m, true
		}
	}
	return mediaModelDef{}, false
}

func normalizeWavespeedImageSlug(model string) string {
	slug := strings.Trim(model, "/")
	if !strings.Contains(slug, "/") {
		slug = "google/" + slug
	}
	if !strings.HasSuffix(slug, "text-to-image") && !strings.HasSuffix(slug, "edit") {
		slug += "/text-to-image"
	}
	return slug
}
