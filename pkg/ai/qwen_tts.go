package ai

import "strings"

const (
	qwen3TTSTextSlug  = "wavespeed-ai/qwen3-tts/text-to-speech"
	qwen3TTSCloneSlug = "wavespeed-ai/qwen3-tts/voice-clone"
)

func isUnifiedQwen3TTS(def mediaModelDef) bool {
	return def.ID == "qwen3-tts"
}

func selectQwen3TTSSlug(req AudioRequest) string {
	if hasQwen3TTSSourceAudio(req) {
		return qwen3TTSCloneSlug
	}
	return qwen3TTSTextSlug
}

func hasQwen3TTSSourceAudio(req AudioRequest) bool {
	if strings.TrimSpace(req.AudioBase64) != "" {
		return true
	}
	if strings.TrimSpace(req.SourceAudioURL) != "" || strings.TrimSpace(req.AudioURL) != "" {
		return true
	}
	return strings.EqualFold(strings.TrimSpace(req.Mode), "clone")
}

func wavespeedAudioModelID(def mediaModelDef, slug string) string {
	if isUnifiedQwen3TTS(def) && slug == qwen3TTSCloneSlug {
		return def.ID + "-clone"
	}
	return def.ID
}
