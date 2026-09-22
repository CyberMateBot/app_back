package billing

// DefaultModelPrice returns the CyberCoin cost for a model operation.
// 1 CyberCoin = 1 ₽. Prices calculated at ~3.5× provider cost.
//
// Reference exchange rate: ~85 ₽/$ (Sep 2026). Provider costs sourced from
// Wavespeed pricing (https://wavespeed.ai/pricing) + OpenAI/Anthropic list.
// For text models the reference workload is ~5000 input + 2000 output tokens.
// For video/audio the reference is the default duration (5 sec / 60 sec for
// ACE-Step / 1 song for Mureka / default text_length for TTS) so that the
// dynamic price scalers in {video,audio}_option_prices.go stay accurate.
func DefaultModelPrice(id, category string) int {
	if price, ok := defaultModelPrices[id]; ok {
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

// defaultModelPrices is the canonical per-model price table (CyberCoins per operation).
// See DefaultModelPrice comment for the methodology; audit performed Sep 2026.
var defaultModelPrices = map[string]int{
	// Text (chat) models — priced against typical ~5k in / ~2k out prompt.
	"yandexgpt": 5, "gpt-oss-20b": 1, "gpt-oss-120b": 3, "qwen3.6-35b": 5, "qwen3-235b": 8,
	"deepseek-v4-flash": 1, "deepseek-chat": 1, "deepseek-v3.2": 3, "deepseek-chat-v3-0324": 1,
	"deepseek-v3.2-exp": 3, "deepseek-v4": 5, "deepseek-r1": 5, "gpt-4.1-nano": 1, "gpt-4o-mini": 1,
	"gpt-5.4-mini": 4, "gpt-4.1-mini": 2, "o4-mini": 4, "gemini-2.5-flash": 2, "claude-haiku-4.5": 5,
	"gpt-4.1": 8, "gpt-4o": 10, "gemini-2.5-pro": 10, "o3-mini": 5, "gpt-5.4": 8, "claude-sonnet-4.5": 14,
	"o3": 39, "o1": 58, "claude-opus-4.7": 22, "claude-opus-4.8": 22, "claude-fable-5": 50, "gpt-5.5": 26,

	// Image models.
	"flux-dev": 5, "nano-banana": 10, "gpt-image-2": 20, "gpt-image-1.5": 20, "nano-banana-2": 22,
	"nano-banana-pro": 42, "alice-ai-art": 12,
	"seedream-v4.5": 14, "seedream-v5.0-lite": 12,
	"qwen-image": 7, "qwen-image-2512": 7, "qwen-image-2.0": 11, "qwen-image-2.0-pro": 25,
	"z-image-base": 5, "z-image-turbo": 2, "grok-imagine-edit": 25,

	// Video models — base = 5-second default; longer durations / higher
	// resolutions are surcharged proportionally in video_option_prices.go.
	"kling": 125, "kling-v3-std": 125, "kling-v3-pro": 166, "kling-v3-4k": 250, "seedance": 149,
	"seedance-v1-pro-i2v": 149, "seedance-v1.5-i2v-fast": 149, "seedance-v1.5-t2v-fast": 149,
	"seedance-v1.5-i2v-spicy": 223, "seedance-v2-video-edit": 223, "seedance-v2-video-extend": 223,
	"wan-2.5-t2v": 149, "wan-2.6-i2v": 149, "wan-2.2-spicy-i2v": 149,
	"wan-2.7-t2v": 149, "wan-2.7-flf": 149, "wan-2.7-grid": 149, "wan-2.7-edit": 223,
	"happyhorse-t2v": 223, "happyhorse-i2v": 223, "happyhorse-ref2v": 223, "happyhorse-video-edit": 223, "happyhorse-video-extend": 223,
	"sora-2-t2v": 149, "sora-2-i2v": 149, "sora-2-t2v-pro": 250,
	"veo-3.1-extend": 312, "vidu-q3-i2v-spicy": 238,
	"hailuo-2.3-t2v": 44, "hailuo-2.3-i2v-fast": 36, "hailuo-2.3-i2v-pro": 93,

	// Audio / TTS — base = default text_length from media_model_options.go,
	// audio_option_prices.go scales for longer inputs.
	"qwen3-tts": 4, "qwen3-tts-clone": 15, "omnivoice": 3, "elevenlabs-v3": 30,
	"minimax-speech-2.6": 18, "mureka": 50, "mureka-v9": 50, "ace-step-1.5": 40,

	// 3D models.
	"hunyuan3d-v3.1-rapid": 25, "hunyuan3d-v3.1-rapid-i2d": 250, "hunyuan3d-v3-t2d": 30, "tripo3d-v2.5-i2d": 48,
	"tripo3d-v2.5-multiview": 48, "tripo3d-h3.1-t2d": 55, "tripo3d-h3.1-i2d": 55,
	"meshy6-t2d": 48, "rodin-v2-i2d": 55, "rodin-v2.5-i2d": 55,
}
