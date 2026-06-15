package ai

func init() {
	wavespeedImageModelCatalog = append(wavespeedImageModelCatalog, extendedImageModels...)
	wavespeedVideoModelCatalog = append(wavespeedVideoModelCatalog, extendedVideoModels...)
	for key, id := range extendedMediaAliases {
		mediaModelAliases[key] = id
	}
}

var extendedImageModels = []mediaModelDef{
	{
		ID: "seedream-v4.5", Label: "Seedream 4.5", Group: "Seedream",
		Description: "ByteDance Seedream 4.5: генерация, редактирование и серии",
		TextSlug: "bytedance/seedream-v4.5", EditSlug: "bytedance/seedream-v4.5/edit",
		MultiSlug: "bytedance/seedream-v4.5/sequential",
		Provider: "wavespeed", Kind: "image",
	},
	{
		ID: "seedream-v5.0-lite", Label: "Seedream 5.0 Lite", Group: "Seedream",
		Description: "Seedream 5.0 Lite до 4K: генерация, edit и sequential",
		TextSlug: "bytedance/seedream-v5.0-lite", EditSlug: "bytedance/seedream-v5.0-lite/edit",
		MultiSlug: "bytedance/seedream-v5.0-lite/sequential",
		EditSequentialSlug: "bytedance/seedream-v5.0-lite/edit-sequential",
		Provider: "wavespeed", Kind: "image",
	},
	{
		ID: "qwen-image", Label: "Qwen Image", Group: "Qwen Image",
		Description: "Qwen Image 20B: постеры и иллюстрации по тексту",
		TextSlug: "wavespeed-ai/qwen-image/text-to-image",
		Provider: "wavespeed", Kind: "image",
	},
	{
		ID: "qwen-image-2512", Label: "Qwen Image 2512", Group: "Qwen Image",
		Description: "Актуальная Qwen Image с negative prompt и seed",
		TextSlug: "wavespeed-ai/qwen-image/text-to-image-2512",
		Provider: "wavespeed", Kind: "image",
	},
	{
		ID: "qwen-image-2.0", Label: "Qwen Image 2.0", Group: "Qwen Image",
		Description: "Qwen Image 2.0 с кастомным размером и пресетами",
		TextSlug: "wavespeed-ai/qwen-image-2.0/text-to-image",
		Provider: "wavespeed", Kind: "image",
	},
	{
		ID: "qwen-image-2.0-pro", Label: "Qwen Image 2.0 Pro Edit", Group: "Qwen Image",
		Description: "Редактирование изображений Qwen Image 2.0 Pro",
		EditSlug: "wavespeed-ai/qwen-image-2.0-pro/edit",
		Provider: "wavespeed", Kind: "image",
	},
	{
		ID: "z-image-base", Label: "Z-Image Base", Group: "Z-Image",
		Description: "Z-Image Base: text-to-image и img2img с strength",
		TextSlug: "wavespeed-ai/z-image/base",
		EditSlug: "wavespeed-ai/z-image/base",
		Provider: "wavespeed", Kind: "image",
	},
	{
		ID: "z-image-turbo", Label: "Z-Image Turbo", Group: "Z-Image",
		Description: "Быстрая Z-Image Turbo генерация",
		TextSlug: "wavespeed-ai/z-image/turbo",
		Provider: "wavespeed", Kind: "image",
	},
	{
		ID: "grok-imagine-edit", Label: "Grok Imagine Edit", Group: "Grok",
		Description: "xAI Grok: редактирование изображений высокого качества",
		EditSlug: "x-ai/grok-imagine-image-quality/edit",
		Provider: "wavespeed", Kind: "image",
	},
}

var extendedVideoModels = []mediaModelDef{
	{
		ID: "wan-2.5-t2v", Label: "WAN 2.5 T2V", Group: "WAN",
		Description: "Alibaba WAN 2.5 text-to-video",
		TextSlug: "alibaba/wan-2.5/text-to-video",
		Provider: "wavespeed", Kind: "video",
	},
	{
		ID: "wan-2.6-i2v", Label: "WAN 2.6 I2V", Group: "WAN",
		Description: "Alibaba WAN 2.6 image-to-video",
		TextSlug: "alibaba/wan-2.6/image-to-video",
		Provider: "wavespeed", Kind: "video", RequiresImage: true,
	},
	{
		ID: "wan-2.7-t2v", Label: "WAN 2.7 T2V", Group: "WAN",
		Description: "Alibaba WAN 2.7 text-to-video до 1080P",
		TextSlug: "alibaba/wan-2.7/text-to-video",
		Provider: "wavespeed", Kind: "video",
	},
	{
		ID: "wan-2.7-flf", Label: "WAN 2.7 First/Last", Group: "WAN",
		Description: "WAN 2.7 переход между первым и последним кадром",
		TextSlug: "alibaba/wan-2.7/first-last-frame",
		Provider: "wavespeed", Kind: "video",
	},
	{
		ID: "wan-2.7-grid", Label: "WAN 2.7 Grid", Group: "WAN",
		Description: "WAN 2.7 анимация 3×3 сетки изображений",
		TextSlug: "alibaba/wan-2.7/image-grid",
		Provider: "wavespeed", Kind: "video", RequiresImage: true,
	},
	{
		ID: "wan-2.7-edit", Label: "WAN 2.7 Edit", Group: "WAN",
		Description: "WAN 2.7 редактирование видео по инструкции",
		TextSlug: "alibaba/wan-2.7/edit",
		Provider: "wavespeed", Kind: "video", RequiresVideo: true,
	},
	{
		ID: "happyhorse-t2v", Label: "Happy Horse T2V", Group: "Happy Horse",
		Description: "Alibaba Happy Horse 1.0 text-to-video",
		TextSlug: "alibaba/happyhorse-1.0/text-to-video",
		Provider: "wavespeed", Kind: "video",
	},
	{
		ID: "happyhorse-i2v", Label: "Happy Horse I2V", Group: "Happy Horse",
		Description: "Happy Horse 1.0 image-to-video",
		TextSlug: "alibaba/happyhorse-1.0/image-to-video",
		Provider: "wavespeed", Kind: "video", RequiresImage: true,
	},
	{
		ID: "happyhorse-ref2v", Label: "Happy Horse Ref2V", Group: "Happy Horse",
		Description: "Happy Horse reference-to-video с сохранением персонажа",
		TextSlug: "alibaba/happyhorse-1.0/reference-to-video",
		Provider: "wavespeed", Kind: "video", RequiresImage: true,
	},
	{
		ID: "happyhorse-video-edit", Label: "Happy Horse Edit", Group: "Happy Horse",
		Description: "Happy Horse редактирование видео",
		TextSlug: "alibaba/happyhorse-1.0/video-edit",
		Provider: "wavespeed", Kind: "video", RequiresVideo: true,
	},
	{
		ID: "happyhorse-video-extend", Label: "Happy Horse Extend", Group: "Happy Horse",
		Description: "Happy Horse продление видео",
		TextSlug: "alibaba/happyhorse-1.0/video-extend",
		Provider: "wavespeed", Kind: "video", RequiresVideo: true,
	},
	{
		ID: "wan-2.2-spicy-i2v", Label: "WAN 2.2 Spicy I2V", Group: "WAN",
		Description: "WAN 2.2 Spicy image-to-video",
		TextSlug: "wavespeed-ai/wan-2.2-spicy/image-to-video",
		Provider: "wavespeed", Kind: "video", RequiresImage: true,
	},
	{
		ID: "sora-2-t2v", Label: "Sora 2 T2V", Group: "Sora",
		Description: "OpenAI Sora 2 text-to-video",
		TextSlug: "openai/sora-2/text-to-video",
		Provider: "wavespeed", Kind: "video",
	},
	{
		ID: "sora-2-i2v", Label: "Sora 2 I2V", Group: "Sora",
		Description: "OpenAI Sora 2 image-to-video",
		TextSlug: "openai/sora-2/image-to-video",
		Provider: "wavespeed", Kind: "video", RequiresImage: true,
	},
	{
		ID: "sora-2-t2v-pro", Label: "Sora 2 Pro T2V", Group: "Sora",
		Description: "OpenAI Sora 2 Pro text-to-video",
		TextSlug: "openai/sora-2/text-to-video-pro",
		Provider: "wavespeed", Kind: "video",
	},
	{
		ID: "veo-3.1-extend", Label: "Veo 3.1 Extend", Group: "Veo",
		Description: "Google Veo 3.1 Fast продление видео",
		TextSlug: "google/veo3.1-fast/video-extend",
		Provider: "wavespeed", Kind: "video", RequiresVideo: true,
	},
	{
		ID: "vidu-q3-i2v-spicy", Label: "Vidu Q3 Spicy I2V", Group: "Vidu",
		Description: "Vidu Q3 image-to-video spicy с аудио",
		TextSlug: "vidu/q3/image-to-video-spicy",
		Provider: "wavespeed", Kind: "video", RequiresImage: true,
	},
	{
		ID: "hailuo-2.3-t2v", Label: "Hailuo 2.3 T2V", Group: "Hailuo",
		Description: "MiniMax Hailuo 2.3 text-to-video",
		TextSlug: "minimax/hailuo-2.3/t2v-standard",
		Provider: "wavespeed", Kind: "video",
	},
	{
		ID: "hailuo-2.3-i2v-fast", Label: "Hailuo 2.3 Fast I2V", Group: "Hailuo",
		Description: "MiniMax Hailuo 2.3 быстрый image-to-video",
		TextSlug: "minimax/hailuo-2.3/fast",
		Provider: "wavespeed", Kind: "video", RequiresImage: true,
	},
	{
		ID: "hailuo-2.3-i2v-pro", Label: "Hailuo 2.3 Pro I2V", Group: "Hailuo",
		Description: "MiniMax Hailuo 2.3 Pro image-to-video 1080p",
		TextSlug: "minimax/hailuo-2.3/i2v-pro",
		Provider: "wavespeed", Kind: "video", RequiresImage: true,
	},
}

var extendedMediaAliases = map[string]string{
	"bytedance/seedream-v4.5":                    "seedream-v4.5",
	"bytedance/seedream-v4.5/edit":               "seedream-v4.5",
	"bytedance/seedream-v4.5/sequential":         "seedream-v4.5",
	"bytedance/seedream-v5.0-lite":               "seedream-v5.0-lite",
	"bytedance/seedream-v5.0-lite/edit":          "seedream-v5.0-lite",
	"bytedance/seedream-v5.0-lite/sequential":    "seedream-v5.0-lite",
	"bytedance/seedream-v5.0-lite/edit-sequential": "seedream-v5.0-lite",
	"wavespeed-ai/qwen-image/text-to-image":      "qwen-image",
	"wavespeed-ai/qwen-image/text-to-image-2512": "qwen-image-2512",
	"wavespeed-ai/qwen-image-2.0/text-to-image":  "qwen-image-2.0",
	"wavespeed-ai/qwen-image-2.0-pro/edit":       "qwen-image-2.0-pro",
	"wavespeed-ai/z-image/base":                  "z-image-base",
	"wavespeed-ai/z-image/turbo":                 "z-image-turbo",
	"x-ai/grok-imagine-image-quality/edit":       "grok-imagine-edit",
	"alibaba/wan-2.5/text-to-video":              "wan-2.5-t2v",
	"alibaba/wan-2.6/image-to-video":             "wan-2.6-i2v",
	"alibaba/wan-2.7/text-to-video":              "wan-2.7-t2v",
	"alibaba/wan-2.7/first-last-frame":           "wan-2.7-flf",
	"alibaba/wan-2.7/image-grid":               "wan-2.7-grid",
	"alibaba/wan-2.7/edit":                     "wan-2.7-edit",
	"alibaba/happyhorse-1.0/text-to-video":     "happyhorse-t2v",
	"alibaba/happyhorse-1.0/image-to-video":    "happyhorse-i2v",
	"alibaba/happyhorse-1.0/reference-to-video": "happyhorse-ref2v",
	"alibaba/happyhorse-1.0/video-edit":        "happyhorse-video-edit",
	"alibaba/happyhorse-1.0/video-extend":      "happyhorse-video-extend",
	"wavespeed-ai/wan-2.2-spicy/image-to-video": "wan-2.2-spicy-i2v",
	"openai/sora-2/text-to-video":              "sora-2-t2v",
	"openai/sora-2/image-to-video":             "sora-2-i2v",
	"openai/sora-2/text-to-video-pro":          "sora-2-t2v-pro",
	"google/veo3.1-fast/video-extend":           "veo-3.1-extend",
	"vidu/q3/image-to-video-spicy":             "vidu-q3-i2v-spicy",
	"minimax/hailuo-2.3/t2v-standard":          "hailuo-2.3-t2v",
	"minimax/hailuo-2.3/fast":                  "hailuo-2.3-i2v-fast",
	"minimax/hailuo-2.3/i2v-pro":               "hailuo-2.3-i2v-pro",
}
