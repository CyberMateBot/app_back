package ai

func init() {
	wavespeedVideoModelCatalog = append(wavespeedVideoModelCatalog, extendedKlingVideoModels...)
	wavespeedImageModelCatalog = append(wavespeedImageModelCatalog, extendedKlingImageModels...)
	wavespeedAudioModelCatalog = append(wavespeedAudioModelCatalog, extendedKlingAudioModels...)
	for key, id := range extendedKlingAliases {
		mediaModelAliases[key] = id
	}
}

var extendedKlingVideoModels = []mediaModelDef{
	{
		ID: "kling-video-o3-std", Label: "Kling Video O3 Standard", Group: "Kling",
		Description: "Kling Omni Video O3 Standard: кинематографичное видео 720p с аудио",
		TextSlug:    "kwaivgi/kling-video-o3-std/text-to-video",
		Provider:    "wavespeed", Kind: "video",
	},
	{
		ID: "kling-video-o3-pro", Label: "Kling Video O3 Pro", Group: "Kling",
		Description: "Kling Omni Video O3 Pro: кинематографичное видео 1080p с аудио",
		TextSlug:    "kwaivgi/kling-video-o3-pro/text-to-video",
		Provider:    "wavespeed", Kind: "video",
	},
	{
		ID: "kling-video-o3-4k", Label: "Kling Video O3 4K", Group: "Kling",
		Description: "Kling Omni Video O3 4K: сверхвысокая четкость 4K с аудио",
		TextSlug:    "kwaivgi/kling-video-o3-4k/text-to-video",
		Provider:    "wavespeed", Kind: "video",
	},
	{
		ID: "kling-v3-turbo-std", Label: "Kling 3.0 Turbo Std", Group: "Kling",
		Description: "Быстрая генерация видео Kling 3.0 Turbo 720p",
		TextSlug:    "kwaivgi/kling-v3-turbo-std/text-to-video",
		Provider:    "wavespeed", Kind: "video",
	},
	{
		ID: "kling-v3-turbo-pro", Label: "Kling 3.0 Turbo Pro", Group: "Kling",
		Description: "Скоростная Pro генерация Kling 3.0 Turbo 1080p",
		TextSlug:    "kwaivgi/kling-v3-turbo-pro/text-to-video",
		Provider:    "wavespeed", Kind: "video",
	},
	{
		ID: "kling-v2.6-std", Label: "Kling 2.6 Standard", Group: "Kling",
		Description: "Плавное движение и кинематографичность Kling 2.6 720p",
		TextSlug:    "kwaivgi/kling-v2.6-std/text-to-video",
		Provider:    "wavespeed", Kind: "video",
	},
	{
		ID: "kling-v2.6-pro", Label: "Kling 2.6 Pro", Group: "Kling",
		Description: "Высокое качество и нативный звук Kling 2.6 1080p",
		TextSlug:    "kwaivgi/kling-v2.6-pro/text-to-video",
		Provider:    "wavespeed", Kind: "video",
	},
	{
		ID: "kling-v2.1-master", Label: "Kling 2.1 Master", Group: "Kling",
		Description: "Кинематографичная динамика Kling 2.1 Master",
		TextSlug:    "kwaivgi/kling-v2.1-t2v-master",
		Provider:    "wavespeed", Kind: "video",
	},
	{
		ID: "kling-v2.0-master", Label: "Kling 2.0 Master", Group: "Kling",
		Description: "Классический Kling 2.0 Master видеогенератор",
		TextSlug:    "kwaivgi/kling-v2.0-t2v-master",
		Provider:    "wavespeed", Kind: "video",
	},
	{
		ID: "kling-v1.6-std", Label: "Kling 1.6 Standard", Group: "Kling",
		Description: "Проверенная генерация Kling 1.6 Standard",
		TextSlug:    "kwaivgi/kling-v1.6-t2v-standard",
		Provider:    "wavespeed", Kind: "video",
	},
	{
		ID: "kling-v1.6-pro", Label: "Kling 1.6 Pro I2V", Group: "Kling",
		Description: "Анимация фото в реалистичное видео Kling 1.6 Pro",
		TextSlug:    "kwaivgi/kling-v1.6-i2v-pro",
		Provider:    "wavespeed", Kind: "video", RequiresImage: true,
	},
	{
		ID: "kling-video-o1", Label: "Kling Video O1", Group: "Kling",
		Description: "Omni видеомодель Kling Video O1",
		TextSlug:    "kwaivgi/kling-video-o1/text-to-video",
		Provider:    "wavespeed", Kind: "video",
	},
}

var extendedKlingImageModels = []mediaModelDef{
	{
		ID: "kling-image-o3", Label: "Kling Image O3", Group: "Kling",
		Description: "Флагманская Kling Image O3: генерация и edit до 4K",
		TextSlug:    "kwaivgi/kling-image-o3/text-to-image",
		EditSlug:    "kwaivgi/kling-image-o3/edit",
		Provider:    "wavespeed", Kind: "image",
	},
	{
		ID: "kling-image-v3", Label: "Kling Image 3.0", Group: "Kling",
		Description: "Высокодетализированная Kling Image 3.0 генерация и edit",
		TextSlug:    "kwaivgi/kling-image-v3/text-to-image",
		EditSlug:    "kwaivgi/kling-image-v3/edit",
		Provider:    "wavespeed", Kind: "image",
	},
}

var extendedKlingAudioModels = []mediaModelDef{
	{
		ID: "kling-v1-tts", Label: "Kling V1 TTS", Group: "Kling",
		Description: "Синтез речи и озвучка Kling V1 TTS",
		TextSlug:    "kwaivgi/kling-v1-tts",
		Provider:    "wavespeed", Kind: "audio",
	},
}

var extendedKlingAliases = map[string]string{
	"kwaivgi/kling-video-o3-std/text-to-video": "kling-video-o3-std",
	"kwaivgi/kling-video-o3-pro/text-to-video": "kling-video-o3-pro",
	"kwaivgi/kling-video-o3-4k/text-to-video":  "kling-video-o3-4k",
	"kwaivgi/kling-v3-turbo-std/text-to-video": "kling-v3-turbo-std",
	"kwaivgi/kling-v3-turbo-pro/text-to-video": "kling-v3-turbo-pro",
	"kwaivgi/kling-v2.6-std/text-to-video":     "kling-v2.6-std",
	"kwaivgi/kling-v2.6-pro/text-to-video":     "kling-v2.6-pro",
	"kwaivgi/kling-v2.1-t2v-master":            "kling-v2.1-master",
	"kwaivgi/kling-v2.0-t2v-master":            "kling-v2.0-master",
	"kwaivgi/kling-v1.6-t2v-standard":          "kling-v1.6-std",
	"kwaivgi/kling-v1.6-i2v-pro":               "kling-v1.6-pro",
	"kwaivgi/kling-video-o1/text-to-video":     "kling-video-o1",
	"kwaivgi/kling-image-o3/text-to-image":     "kling-image-o3",
	"kwaivgi/kling-image-o3/edit":              "kling-image-o3",
	"kwaivgi/kling-image-v3/text-to-image":     "kling-image-v3",
	"kwaivgi/kling-image-v3/edit":              "kling-image-v3",
	"kwaivgi/kling-v1-tts":                     "kling-v1-tts",
}
