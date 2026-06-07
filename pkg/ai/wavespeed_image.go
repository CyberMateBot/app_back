package ai

import (
	"context"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func generateWavespeedImage(ctx context.Context, cfg config.ConfigAI, prompt string, req ImageRequest, def mediaModelDef) (ImageResponse, error) {
	slug, err := resolveWavespeedImageSlug(def, req)
	if err != nil {
		return ImageResponse{}, err
	}

	enrichedPrompt := enrichMediaPrompt(prompt, req.Messages)
	input := buildWavespeedImageInput(cfg, def, enrichedPrompt, req, slug)
	urls, err := runWavespeedModel(ctx, cfg, slug, input)
	if err != nil {
		return ImageResponse{}, err
	}

	resp := ImageResponse{
		ImageURL:  urls[0],
		ImageURLs: urls,
		Model:     def.ID,
	}
	if strings.Contains(slug, "/edit") {
		resp.Model = def.ID + "-edit"
	} else if strings.Contains(slug, "multi") {
		resp.Model = def.ID + "-multi"
	}
	return resp, nil
}

func buildWavespeedImageInput(cfg config.ConfigAI, def mediaModelDef, prompt string, req ImageRequest, slug string) map[string]any {
	input := map[string]any{
		"enable_sync_mode":     cfg.NanoBananaSyncMode,
		"enable_base64_output": cfg.NanoBananaBase64Out,
	}

	isEdit := strings.HasSuffix(slug, "/edit")
	isMulti := strings.Contains(slug, "text-to-image-multi")

	if isEdit {
		sourceURL := strings.TrimSpace(req.SourceImageURL)
		if sourceURL == "" {
			sourceURL = strings.TrimSpace(req.ImageURL)
		}
		input["image"] = sourceURL
		input["prompt"] = prompt
	} else {
		input["prompt"] = prompt
	}

	outputFmt := strings.TrimSpace(req.OutputFormat)
	if outputFmt == "" {
		outputFmt = cfg.NanoBananaOutputFmt
	}
	if outputFmt == "" {
		outputFmt = "png"
	}
	input["output_format"] = outputFmt

	if ar := strings.TrimSpace(req.AspectRatio); ar != "" {
		input["aspect_ratio"] = ar
	} else if !isEdit {
		input["aspect_ratio"] = defaultAspectRatio(def.ID)
	}

	resolution := strings.TrimSpace(req.Resolution)
	if resolution == "" && def.ID != "nano-banana" {
		resolution = cfg.NanoBananaResolution
	}
	if resolution != "" && !isEdit {
		input["resolution"] = resolution
	}

	if q := strings.TrimSpace(req.Quality); q != "" {
		input["quality"] = q
	} else if modelSupportsQuality(def.ID) {
		input["quality"] = "medium"
	}

	if isMulti {
		num := req.NumImages
		if num <= 0 {
			num = 2
		}
		if num > 4 {
			num = 4
		}
		input["num_images"] = num
	}

	if size := strings.TrimSpace(req.Size); size != "" {
		input["size"] = size
	}

	return input
}

func defaultAspectRatio(modelID string) string {
	switch modelID {
	case "nano-banana-pro", "nano-banana-2", "gpt-image-2":
		return "16:9"
	default:
		return "1:1"
	}
}
