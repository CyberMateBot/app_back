package ai

import (
	"context"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func generateWavespeedAudio(ctx context.Context, cfg config.ConfigAI, prompt string, req AudioRequest, def mediaModelDef) (AudioResponse, error) {
	if err := prepareWavespeedAudioSource(ctx, cfg, &req); err != nil {
		return AudioResponse{}, err
	}

	slug := selectQwen3TTSSlug(req)
	input := buildWavespeedAudioInput(prompt, req, slug)
	urls, err := runWavespeedModel(ctx, cfg, slug, input)
	if err != nil {
		return AudioResponse{}, err
	}

	return AudioResponse{
		AudioURL: urls[0],
		Model:    wavespeedAudioModelID(def, slug),
	}, nil
}

func buildWavespeedAudioInput(prompt string, req AudioRequest, slug string) map[string]any {
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
