package ai

import (
	"context"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func generateWavespeedImage(ctx context.Context, cfg config.ConfigAI, prompt string, req ImageRequest, def mediaModelDef) (ImageResponse, error) {
	slug := def.Slug
	if def.ID == "nano-banana" && strings.TrimSpace(cfg.NanoBananaModel) != "" {
		slug = normalizeWavespeedImageSlug(cfg.NanoBananaModel)
	}

	input := buildWavespeedImageInput(cfg, def, prompt, req)
	urls, err := runWavespeedModel(ctx, cfg, slug, input)
	if err != nil {
		return ImageResponse{}, err
	}

	return ImageResponse{ImageURL: urls[0], Model: def.ID}, nil
}

func buildWavespeedImageInput(cfg config.ConfigAI, def mediaModelDef, prompt string, req ImageRequest) map[string]any {
	input := map[string]any{"prompt": prompt}

	if def.ID == "nano-banana" {
		if cfg.NanoBananaResolution != "" {
			input["resolution"] = cfg.NanoBananaResolution
		}
		if cfg.NanoBananaOutputFmt != "" {
			input["output_format"] = cfg.NanoBananaOutputFmt
		}
		input["enable_sync_mode"] = cfg.NanoBananaSyncMode
		input["enable_base64_output"] = cfg.NanoBananaBase64Out
	}

	if ar := strings.TrimSpace(req.AspectRatio); ar != "" {
		input["aspect_ratio"] = ar
	}
	if size := strings.TrimSpace(req.Size); size != "" {
		input["size"] = size
	}

	return input
}
