package ai

import (
	"context"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func generateWavespeedVideo(ctx context.Context, cfg config.ConfigAI, prompt string, req VideoRequest, def mediaModelDef) (VideoResponse, error) {
	input := buildWavespeedVideoInput(prompt, req)
	urls, err := runWavespeedModel(ctx, cfg, def.Slug, input)
	if err != nil {
		return VideoResponse{}, err
	}

	return VideoResponse{VideoURL: urls[0], Model: def.ID}, nil
}

func buildWavespeedVideoInput(prompt string, req VideoRequest) map[string]any {
	input := map[string]any{"prompt": prompt}

	duration := req.Duration
	if duration <= 0 {
		duration = 5
	}
	input["duration"] = duration

	if ar := strings.TrimSpace(req.AspectRatio); ar != "" {
		input["aspect_ratio"] = ar
	} else {
		input["aspect_ratio"] = "16:9"
	}

	return input
}
