package ai

import (
	"context"
	"math"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func generateWavespeedAudio(ctx context.Context, cfg config.ConfigAI, prompt string, req AudioRequest, def mediaModelDef) (AudioResponse, error) {
	if err := prepareWavespeedAudioSource(ctx, cfg, &req); err != nil {
		return AudioResponse{}, err
	}

	slug := resolveWavespeedAudioSlug(def, req)
	input := buildWavespeedAudioInput(def, prompt, req, slug)
	urls, err := runWavespeedModel(ctx, cfg, slug, input)
	if err != nil {
		return AudioResponse{}, err
	}

	return AudioResponse{
		AudioURL: urls[0],
		Model:    wavespeedAudioModelID(def, slug),
	}, nil
}

func resolveWavespeedAudioSlug(def mediaModelDef, req AudioRequest) string {
	if isUnifiedQwen3TTS(def) && hasQwen3TTSSourceAudio(req) {
		return qwen3TTSCloneSlug
	}
	return def.TextSlug
}

func buildWavespeedAudioInput(def mediaModelDef, prompt string, req AudioRequest, slug string) map[string]any {
	if isUnifiedQwen3TTS(def) {
		return buildQwen3TTSInput(prompt, req, slug)
	}

	switch def.ID {
	case "omnivoice":
		return buildOmniVoiceInput(prompt, req)
	case "elevenlabs-v3":
		return buildElevenLabsInput(prompt, req)
	case "minimax-speech-2.6":
		return buildMiniMaxSpeechInput(prompt, req)
	case "mureka-v9":
		return buildMurekaInput(prompt, req)
	case "ace-step-1.5":
		return buildACEStepInput(prompt, req)
	default:
		return map[string]any{"text": prompt}
	}
}

func buildQwen3TTSInput(prompt string, req AudioRequest, slug string) map[string]any {
	isClone := strings.HasSuffix(slug, "/voice-clone")

	if isClone {
		sourceURL := strings.TrimSpace(req.SourceAudioURL)
		if sourceURL == "" {
			sourceURL = strings.TrimSpace(req.AudioURL)
		}
		input := map[string]any{
			"audio": sourceURL,
			"text":  prompt,
		}
		if ref := strings.TrimSpace(req.ReferenceText); ref != "" {
			input["reference_text"] = ref
		}
		if lang := strings.TrimSpace(req.Language); lang != "" {
			input["language"] = lang
		} else {
			input["language"] = "auto"
		}
		return input
	}

	input := map[string]any{
		"text": prompt,
	}
	if lang := strings.TrimSpace(req.Language); lang != "" {
		input["language"] = lang
	} else {
		input["language"] = "auto"
	}
	voice := strings.TrimSpace(req.Voice)
	if voice == "" {
		voice = "Dylan"
	}
	input["voice"] = voice
	if style := strings.TrimSpace(req.StyleInstruction); style != "" {
		input["style_instruction"] = style
	}
	return input
}

func buildOmniVoiceInput(prompt string, req AudioRequest) map[string]any {
	input := map[string]any{
		"text":  prompt,
		"speed": normalizeAudioSpeed(req.Speed, 1.0),
	}
	if desc := strings.TrimSpace(req.StyleInstruction); desc != "" {
		input["voice_description"] = desc
	}
	return input
}

func buildElevenLabsInput(prompt string, req AudioRequest) map[string]any {
	voice := strings.TrimSpace(req.Voice)
	if voice == "" {
		voice = "Alice"
	}
	similarity := req.Similarity
	if similarity <= 0 {
		similarity = 1.0
	}
	stability := req.Stability
	if stability <= 0 {
		stability = 0.5
	}
	input := map[string]any{
		"text":       prompt,
		"voice_id":   voice,
		"similarity": similarity,
		"stability":  stability,
	}
	if req.UseSpeakerBoost != nil {
		input["use_speaker_boost"] = *req.UseSpeakerBoost
	} else {
		input["use_speaker_boost"] = true
	}
	return input
}

func buildMiniMaxSpeechInput(prompt string, req AudioRequest) map[string]any {
	voice := strings.TrimSpace(req.Voice)
	if voice == "" {
		voice = "Friendly_Person"
	}
	emotion := strings.TrimSpace(req.Emotion)
	if emotion == "" {
		emotion = "happy"
	}
	format := strings.TrimSpace(req.Format)
	if format == "" {
		format = "mp3"
	}
	langBoost := strings.TrimSpace(req.LanguageBoost)
	if langBoost == "" {
		langBoost = strings.TrimSpace(req.Language)
	}
	if langBoost == "" {
		langBoost = "en"
	}

	input := map[string]any{
		"text":                  prompt,
		"voice_id":              voice,
		"emotion":               emotion,
		"speed":                 normalizeAudioSpeed(req.Speed, 1.0),
		"pitch":                 req.Pitch,
		"volume":                normalizeAudioVolume(req.Volume, 1.0),
		"format":                format,
		"channel":               "1",
		"sample_rate":           44100,
		"bitrate":               128000,
		"language_boost":        langBoost,
		"english_normalization": req.EnglishNormalization == nil || *req.EnglishNormalization,
	}
	return input
}

func buildMurekaInput(prompt string, req AudioRequest) map[string]any {
	style := strings.TrimSpace(req.StyleInstruction)
	if style == "" {
		style = strings.TrimSpace(req.Tags)
	}
	if style == "" {
		style = "upbeat pop, energetic"
	}
	numberOfSongs := req.NumberOfSongs
	if numberOfSongs < 1 {
		numberOfSongs = 1
	}
	if numberOfSongs > 3 {
		numberOfSongs = 3
	}
	outputFormat := strings.TrimSpace(req.OutputFormat)
	if outputFormat == "" {
		outputFormat = "mp3"
	}
	return map[string]any{
		"lyrics":            prompt,
		"prompt":            style,
		"number_of_songs":   numberOfSongs,
		"output_format":     outputFormat,
	}
}

func buildACEStepInput(prompt string, req AudioRequest) map[string]any {
	tags := strings.TrimSpace(req.Tags)
	if tags == "" {
		tags = strings.TrimSpace(req.StyleInstruction)
	}
	duration := req.Duration
	if duration < 5 {
		duration = 60
	}
	if duration > 240 {
		duration = 240
	}
	seed := req.Seed
	if seed == 0 {
		seed = -1
	}
	input := map[string]any{
		"tags":     tags,
		"duration": duration,
		"seed":     seed,
	}
	if prompt != "" {
		input["lyrics"] = prompt
	} else {
		input["lyrics"] = ""
	}
	return input
}

func normalizeAudioSpeed(speed, fallback float64) float64 {
	if speed <= 0 {
		return fallback
	}
	return math.Min(5.0, math.Max(0.1, speed))
}

func normalizeAudioVolume(volume, fallback float64) float64 {
	if volume <= 0 {
		return fallback
	}
	return math.Min(10.0, math.Max(0.1, volume))
}
