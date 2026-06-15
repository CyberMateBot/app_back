package ai

import (
	"context"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func generateWavespeedImage(ctx context.Context, cfg config.ConfigAI, prompt string, req ImageRequest, def mediaModelDef) (ImageResponse, error) {
	if err := prepareWavespeedImageSource(ctx, cfg, &req); err != nil {
		return ImageResponse{}, err
	}

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

	isEdit := strings.HasSuffix(slug, "/edit") || strings.Contains(slug, "/edit-")
	isMulti := strings.Contains(slug, "multi") || strings.Contains(slug, "sequential")

	sourceURL := strings.TrimSpace(req.SourceImageURL)
	if sourceURL == "" {
		sourceURL = strings.TrimSpace(req.ImageURL)
	}

	if isEdit || (sourceURL != "" && def.ID != "z-image-base") {
		applyWavespeedImageSourceInput(input, def.ID, slug, sourceURL, req)
		input["prompt"] = prompt
	} else if sourceURL != "" && def.ID == "z-image-base" {
		input["prompt"] = prompt
		applyWavespeedImageSourceInput(input, def.ID, slug, sourceURL, req)
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
	if def.ID == "seedream-v5.0-lite" && outputFmt == "" {
		outputFmt = "jpeg"
	}
	input["output_format"] = normalizeImageOutputFormat(def.ID, outputFmt)

	if ar := strings.TrimSpace(req.AspectRatio); ar != "" && !wavespeedImageUsesSizeField(def.ID) {
		input["aspect_ratio"] = ar
	} else if !isEdit && !wavespeedImageUsesSizeField(def.ID) {
		input["aspect_ratio"] = defaultAspectRatio(def.ID)
	}

	resolution := strings.TrimSpace(req.Resolution)
	if resolution == "" && def.ID != "nano-banana" {
		resolution = cfg.NanoBananaResolution
	}
	if resolution != "" && modelSupportsResolution(def.ID) {
		if norm := normalizeImageResolution(def.ID, resolution); norm != "" {
			input["resolution"] = norm
		}
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

	if size := strings.TrimSpace(req.Size); size != "" && !wavespeedImageUsesSizeField(def.ID) {
		input["size"] = size
	}

	applyWavespeedImageExtraFields(input, def.ID, req)

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
