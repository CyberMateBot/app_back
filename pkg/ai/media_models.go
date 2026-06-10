package ai

import "strings"

// MediaModel describes a selectable image or video generation model.
type MediaModel struct {
	ID            string        `json:"id"`
	Label         string        `json:"label"`
	Group         string        `json:"group"`
	Description   string        `json:"description,omitempty"`
	Provider      string        `json:"provider"` // wavespeed | yandex
	Kind          string        `json:"kind"`     // image | video | audio
	SupportsEdit   bool          `json:"supports_edit,omitempty"`
	SupportsMulti  bool          `json:"supports_multi,omitempty"`
	RequiresImage  bool          `json:"requires_image,omitempty"`
	RequiresVideo  bool          `json:"requires_video,omitempty"`
	Options        []MediaOption `json:"options,omitempty"`
}

type mediaModelDef struct {
	ID            string
	Label         string
	Group         string
	Description   string
	TextSlug      string
	EditSlug      string
	MultiSlug     string
	Provider      string
	Kind          string
	RequiresImage bool
	RequiresVideo bool
}

var wavespeedImageModelCatalog = []mediaModelDef{
	{
		ID: "nano-banana", Label: "Nano Banana", Group: "Nano Banana",
		Description: "Быстрая генерация и редактирование изображений",
		TextSlug: "google/nano-banana/text-to-image",
		EditSlug: "google/nano-banana/edit",
		Provider: "wavespeed", Kind: "image",
	},
	{
		ID: "nano-banana-pro", Label: "Nano Banana Pro", Group: "Nano Banana",
		Description: "Высокое качество, multi-image и редактирование",
		TextSlug: "google/nano-banana-pro/text-to-image",
		EditSlug: "google/nano-banana-pro/edit",
		MultiSlug: "google/nano-banana-pro/text-to-image-multi",
		Provider: "wavespeed", Kind: "image",
	},
	{
		ID: "nano-banana-2", Label: "Nano Banana 2", Group: "Nano Banana",
		Description: "Новейшая модель Google с 4K и редактированием",
		TextSlug: "google/nano-banana-2/text-to-image",
		EditSlug: "google/nano-banana-2/edit",
		Provider: "wavespeed", Kind: "image",
	},
	{
		ID: "gpt-image-2", Label: "GPT Image 2.0", Group: "GPT Image",
		Description: "OpenAI GPT Image 2.0 — генерация и редактирование",
		TextSlug: "openai/gpt-image-2/text-to-image",
		EditSlug: "openai/gpt-image-2/edit",
		Provider: "wavespeed", Kind: "image",
	},
	{
		ID: "gpt-image-1.5", Label: "GPT Image 1.5", Group: "GPT Image",
		Description: "OpenAI GPT Image 1.5 — генерация и редактирование",
		TextSlug: "openai/gpt-image-1.5/text-to-image",
		EditSlug: "openai/gpt-image-1.5/edit",
		Provider: "wavespeed", Kind: "image",
	},
	{
		ID: "flux-dev", Label: "FLUX Dev", Group: "WaveSpeed",
		Description: "Качественные изображения по текстовому описанию",
		TextSlug: "wavespeed-ai/flux-dev",
		Provider: "wavespeed", Kind: "image",
	},
}

var wavespeedVideoModelCatalog = []mediaModelDef{
	{
		ID: "kling-v3-std", Label: "Kling 3.0 Standard", Group: "Kling",
		Description: "Быстрая генерация видео по тексту",
		TextSlug: "kwaivgi/kling-v3.0-std/text-to-video",
		Provider: "wavespeed", Kind: "video",
	},
	{
		ID: "kling-v3-pro", Label: "Kling 3.0 Pro", Group: "Kling",
		Description: "Высокое качество видео по тексту",
		TextSlug: "kwaivgi/kling-v3.0-pro/text-to-video",
		Provider: "wavespeed", Kind: "video",
	},
	{
		ID: "kling-v3-4k", Label: "Kling 3.0 4K", Group: "Kling",
		Description: "Видео в разрешении 4K",
		TextSlug: "kwaivgi/kling-v3.0-4k/text-to-video",
		Provider: "wavespeed", Kind: "video",
	},
	{
		ID: "seedance-v1-pro-i2v", Label: "Seedance 1.0 I2V", Group: "Seedance",
		Description: "Image-to-video 720p от ByteDance",
		TextSlug: "bytedance/seedance-v1-pro-i2v-720p",
		Provider: "wavespeed", Kind: "video", RequiresImage: true,
	},
	{
		ID: "seedance-v1.5-i2v-fast", Label: "Seedance 1.5 I2V Fast", Group: "Seedance",
		Description: "Быстрый image-to-video с аудио",
		TextSlug: "bytedance/seedance-v1.5-pro/image-to-video-fast",
		Provider: "wavespeed", Kind: "video", RequiresImage: true,
	},
	{
		ID: "seedance-v1.5-t2v-fast", Label: "Seedance 1.5 T2V Fast", Group: "Seedance",
		Description: "Быстрый text-to-video с аудио",
		TextSlug: "bytedance/seedance-v1.5-pro/text-to-video-fast",
		Provider: "wavespeed", Kind: "video",
	},
	{
		ID: "seedance-v1.5-i2v-spicy", Label: "Seedance 1.5 I2V Spicy", Group: "Seedance",
		Description: "Выразительный image-to-video с динамикой",
		TextSlug: "bytedance/seedance-v1.5-pro/image-to-video-spicy",
		Provider: "wavespeed", Kind: "video", RequiresImage: true,
	},
	{
		ID: "seedance-v2-video-edit", Label: "Seedance 2.0 Edit", Group: "Seedance",
		Description: "Редактирование видео: 480p — стандарт, 720p/1080p — Turbo",
		TextSlug: seedanceVideoEditSlug,
		EditSlug: seedanceVideoEditTurboSlug,
		Provider: "wavespeed", Kind: "video", RequiresVideo: true,
	},
	{
		ID: "seedance-v2-video-extend", Label: "Seedance 2.0 Extend", Group: "Seedance",
		Description: "Продление видео новым сегментом",
		TextSlug: "bytedance/seedance-2.0/video-extend",
		Provider: "wavespeed", Kind: "video", RequiresVideo: true,
	},
}

var wavespeedAudioModelCatalog = []mediaModelDef{
	{
		ID: "qwen3-tts", Label: "Qwen3 TTS", Group: "Qwen3 TTS",
		Description: "Озвучка текста и клонирование голоса",
		TextSlug: qwen3TTSTextSlug,
		EditSlug: qwen3TTSCloneSlug,
		Provider: "wavespeed", Kind: "audio",
	},
}

var mediaModelAliases = map[string]string{
	"nano-banana": "nano-banana", "banana": "nano-banana",
	"google/nano-banana/text-to-image": "nano-banana",
	"google/nano-banana/edit":            "nano-banana",
	"nano-banana-pro": "nano-banana-pro",
	"google/nano-banana-pro/text-to-image":       "nano-banana-pro",
	"google/nano-banana-pro/edit":                "nano-banana-pro",
	"google/nano-banana-pro/text-to-image-multi": "nano-banana-pro",
	"nano-banana-2": "nano-banana-2",
	"google/nano-banana-2/text-to-image": "nano-banana-2",
	"google/nano-banana-2/edit":          "nano-banana-2",
	"gpt-image-2": "gpt-image-2",
	"openai/gpt-image-2/text-to-image":  "gpt-image-2",
	"openai/gpt-image-2/edit":            "gpt-image-2",
	"gpt-image-1.5": "gpt-image-1.5",
	"openai/gpt-image-1.5/text-to-image": "gpt-image-1.5",
	"openai/gpt-image-1.5/edit":          "gpt-image-1.5",
	"flux-dev": "flux-dev", "wavespeed-ai/flux-dev": "flux-dev",
	"kling-v3-std": "kling-v3-std",
	"kwaivgi/kling-v3.0-std/text-to-video": "kling-v3-std",
	"kling-v3-pro": "kling-v3-pro",
	"kwaivgi/kling-v3.0-pro/text-to-video": "kling-v3-pro",
	"kling-v3-4k": "kling-v3-4k",
	"kwaivgi/kling-v3.0-4k/text-to-video": "kling-v3-4k",
	"seedance-v1-pro-i2v": "seedance-v1-pro-i2v",
	"bytedance/seedance-v1-pro-i2v-720p": "seedance-v1-pro-i2v",
	"seedance-v1.5-i2v-fast": "seedance-v1.5-i2v-fast",
	"bytedance/seedance-v1.5-pro/image-to-video-fast": "seedance-v1.5-i2v-fast",
	"seedance-v1.5-t2v-fast": "seedance-v1.5-t2v-fast",
	"bytedance/seedance-v1.5-pro/text-to-video-fast": "seedance-v1.5-t2v-fast",
	"seedance-v1.5-i2v-spicy": "seedance-v1.5-i2v-spicy",
	"bytedance/seedance-v1.5-pro/image-to-video-spicy": "seedance-v1.5-i2v-spicy",
	"seedance-v2-video-edit": "seedance-v2-video-edit",
	"bytedance/seedance-2.0/video-edit": "seedance-v2-video-edit",
	"seedance-v2-video-edit-turbo": "seedance-v2-video-edit",
	"bytedance/seedance-2.0/video-edit-turbo": "seedance-v2-video-edit",
	"seedance-v2-video-extend": "seedance-v2-video-extend",
	"bytedance/seedance-2.0/video-extend": "seedance-v2-video-extend",
	"qwen3-tts": "qwen3-tts",
	"wavespeed-ai/qwen3-tts/text-to-speech": "qwen3-tts",
	"wavespeed-ai/qwen3-tts/voice-clone":  "qwen3-tts",
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

func ListAudioModels() []MediaModel {
	out := make([]MediaModel, 0, len(wavespeedAudioModelCatalog))
	for _, m := range wavespeedAudioModelCatalog {
		out = append(out, toMediaModel(m))
	}
	return out
}

func toMediaModel(m mediaModelDef) MediaModel {
	opts := imageModelOptions(m.ID)
	switch m.Kind {
	case "video":
		opts = videoModelOptions(m.ID)
	case "audio":
		opts = audioModelOptions(m.ID)
	}
	return MediaModel{
		ID: m.ID, Label: m.Label, Group: m.Group,
		Description: m.Description, Provider: m.Provider, Kind: m.Kind,
		SupportsEdit: m.EditSlug != "" || m.RequiresVideo,
		SupportsMulti: m.MultiSlug != "",
		RequiresImage: m.RequiresImage,
		RequiresVideo: m.RequiresVideo,
		Options: opts,
	}
}

func resolveWavespeedVideoSlug(def mediaModelDef, req VideoRequest) (string, error) {
	if def.TextSlug == "" {
		return "", &ProviderError{Provider: "wavespeed", Message: "unknown video model slug"}
	}

	sourceVideo := strings.TrimSpace(req.SourceVideoURL)
	if sourceVideo == "" {
		sourceVideo = strings.TrimSpace(req.VideoURL)
	}
	sourceImage := strings.TrimSpace(req.SourceImageURL)
	if sourceImage == "" {
		sourceImage = strings.TrimSpace(req.ImageURL)
	}

	if def.RequiresVideo && sourceVideo == "" {
		return "", &ProviderError{Provider: "wavespeed", Message: "source video is required for this model"}
	}
	if def.RequiresImage && sourceImage == "" {
		return "", &ProviderError{Provider: "wavespeed", Message: "source image is required for this model"}
	}

	if isUnifiedSeedanceVideoEdit(def) {
		return selectSeedanceVideoEditSlug(req), nil
	}

	return def.TextSlug, nil
}

func resolveWavespeedImageModel(requested string) (mediaModelDef, bool) {
	return resolveMediaModel(wavespeedImageModelCatalog, requested)
}

func resolveWavespeedVideoModel(requested string) (mediaModelDef, bool) {
	return resolveMediaModel(wavespeedVideoModelCatalog, requested)
}

func resolveWavespeedAudioModel(requested string) (mediaModelDef, bool) {
	return resolveMediaModel(wavespeedAudioModelCatalog, requested)
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
		if m.ID == key {
			return m, true
		}
		if strings.EqualFold(m.TextSlug, key) || strings.EqualFold(m.EditSlug, key) || strings.EqualFold(m.MultiSlug, key) {
			return m, true
		}
	}
	return mediaModelDef{}, false
}

func resolveWavespeedImageSlug(def mediaModelDef, req ImageRequest) (string, error) {
	sourceURL := strings.TrimSpace(req.SourceImageURL)
	if sourceURL == "" {
		sourceURL = strings.TrimSpace(req.ImageURL)
	}

	mode := strings.ToLower(strings.TrimSpace(req.Mode))
	if sourceURL != "" || mode == "edit" {
		if def.EditSlug == "" {
			return "", &ProviderError{Provider: "wavespeed", Message: "model does not support image editing"}
		}
		if sourceURL == "" {
			return "", &ProviderError{Provider: "wavespeed", Message: "source image is required for edit mode"}
		}
		return def.EditSlug, nil
	}

	if mode == "multi" {
		if def.MultiSlug == "" {
			return "", &ProviderError{Provider: "wavespeed", Message: "model does not support multi-image generation"}
		}
		return def.MultiSlug, nil
	}

	if def.TextSlug != "" {
		return def.TextSlug, nil
	}
	return "", &ProviderError{Provider: "wavespeed", Message: "unknown image model slug"}
}
