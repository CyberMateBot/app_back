package config

import (
	"os"
	"strings"
	"time"
)

// ConfigAI holds credentials and defaults for AI generation providers.
type ConfigAI struct {
	YandexAPIKey     string
	YandexFolderID   string
	YandexGPTModel   string
	YandexDeepSeek   string
	YandexImageModel string
	YandexImageSize  string

	OpenAIAPIKey  string
	OpenAITextModel string

	WavespeedAPIKey       string
	GeminiAPIBaseURL      string
	GeminiModel           string
	NanoBananaModel       string
	NanoBananaResolution  string
	NanoBananaOutputFmt   string
	NanoBananaSyncMode    bool
	NanoBananaBase64Out   bool
	NanoBananaPollMS      int

	TextMaxOutputTokens int
	HTTPTimeout         time.Duration
	ImagePollTimeout    time.Duration
}

func LoadAIConfig() ConfigAI {
	pollMS := getenvInt("NANO_BANANA_POLL_MS", 600)
	if pollMS < 200 {
		pollMS = 200
	}

	return ConfigAI{
		YandexAPIKey:     strings.TrimSpace(os.Getenv("YANDEX_GPT_API_KEY")),
		YandexFolderID:   strings.TrimSpace(os.Getenv("YANDEX_GPT_FOLDER_ID")),
		YandexGPTModel:   getenv("YANDEX_GPT_MODEL", "yandexgpt/latest"),
		YandexDeepSeek:   getenv("YANDEX_DEEPSEEK_MODEL", "deepseek-v32/latest"),
		// For Alice AI ART via OpenAI-compatible Images API we keep /latest by default.
		YandexImageModel: getenv("YANDEX_ALICE_AI_ART_MODEL", "aliceai-image-art-3.0/latest"),
		YandexImageSize:  getenv("YANDEX_IMAGE_SIZE", "1024x1024"),

		OpenAIAPIKey:    strings.TrimSpace(os.Getenv("OPENAI_API_KEY")),
		OpenAITextModel: getenv("OPENAI_TEXT_MODEL", "gpt-4o-mini"),

		WavespeedAPIKey:      strings.TrimSpace(os.Getenv("WAVESPEED_API_KEY")),
		GeminiAPIBaseURL:     strings.TrimRight(getenv("GEMINI_API_BASE_URL", "https://llm.wavespeed.ai/v1"), "/"),
		GeminiModel:          getenv("GEMINI_MODEL", "google/gemini-2.5-flash"),
		NanoBananaModel:      getenv("NANO_BANANA_MODEL", "google/nano-banana-pro"),
		NanoBananaResolution: getenv("NANO_BANANA_RESOLUTION", "1k"),
		NanoBananaOutputFmt:  getenv("NANO_BANANA_OUTPUT_FORMAT", "jpeg"),
		NanoBananaSyncMode:   getenvBool("NANO_BANANA_SYNC_MODE", false),
		NanoBananaBase64Out:  getenvBool("NANO_BANANA_BASE64_OUTPUT", false),
		NanoBananaPollMS:     pollMS,

		TextMaxOutputTokens: getenvInt("AI_TEXT_MAX_OUTPUT_TOKENS", 4096),
		HTTPTimeout:         getenvDuration("AI_HTTP_TIMEOUT", 120*time.Second),
		ImagePollTimeout:    getenvDuration("AI_IMAGE_POLL_TIMEOUT", 3*time.Minute),
	}
}

func (c ConfigAI) YandexTextEnabled() bool {
	return c.YandexAPIKey != "" && c.YandexFolderID != ""
}

func (c ConfigAI) OpenAITextEnabled() bool {
	return c.OpenAIAPIKey != "" && c.OpenAIAPIKey != "..."
}

func (c ConfigAI) WavespeedTextEnabled() bool {
	return c.WavespeedAPIKey != ""
}

func (c ConfigAI) WavespeedImageEnabled() bool {
	return c.WavespeedAPIKey != ""
}
