package ai

import (
	"strconv"
	"strings"
)

func wavespeedImageUsesSingularImage(modelID string) bool {
	switch modelID {
	case "seedream-v4.5", "seedream-v5.0-lite", "qwen-image", "qwen-image-2512",
		"z-image-base", "z-image-turbo", "grok-imagine-edit":
		return true
	default:
		return false
	}
}

func wavespeedImageUsesPluralImages(modelID string) bool {
	return modelID == "qwen-image-2.0-pro"
}

func wavespeedImageUsesSizeField(modelID string) bool {
	switch modelID {
	case "qwen-image", "qwen-image-2512", "qwen-image-2.0", "z-image-base", "z-image-turbo",
		"seedream-v5.0-lite":
		return true
	default:
		return false
	}
}

func wavespeedImageSizeFromRequest(modelID string, req ImageRequest) string {
	if size := strings.TrimSpace(req.Size); size != "" {
		return size
	}
	if ar := strings.TrimSpace(req.AspectRatio); ar != "" && (modelID == "qwen-image-2.0" || strings.HasPrefix(modelID, "z-image")) {
		return ar
	}
	if req.Width > 0 && req.Height > 0 {
		return formatPixelSize(req.Width, req.Height)
	}
	return ""
}

func formatPixelSize(width, height int) string {
	return strconv.Itoa(width) + "*" + strconv.Itoa(height)
}

func applyWavespeedImageSourceInput(input map[string]any, modelID, slug, sourceURL string, req ImageRequest) {
	if sourceURL == "" {
		return
	}

	switch {
	case wavespeedImageUsesPluralImages(modelID):
		input["images"] = []string{sourceURL}
	case wavespeedImageUsesSingularImage(modelID):
		input["image"] = sourceURL
	default:
		input["images"] = []string{sourceURL}
	}

	if modelID == "z-image-base" && !strings.HasSuffix(slug, "/edit") && req.Strength > 0 {
		input["strength"] = req.Strength
	}
}

func applyWavespeedImageExtraFields(input map[string]any, modelID string, req ImageRequest) {
	if neg := strings.TrimSpace(req.NegativePrompt); neg != "" {
		input["negative_prompt"] = neg
	}

	if req.Seed != 0 {
		input["seed"] = req.Seed
	} else if modelID == "qwen-image-2512" || strings.HasPrefix(modelID, "z-image") || modelID == "qwen-image-2.0" {
		input["seed"] = -1
	}

	if req.PromptExtend != nil && modelID == "qwen-image-2512" {
		input["prompt_extend"] = *req.PromptExtend
	}

	if size := wavespeedImageSizeFromRequest(modelID, req); size != "" && wavespeedImageUsesSizeField(modelID) {
		input["size"] = size
	}

	if modelID == "grok-imagine-edit" {
		if ar := strings.TrimSpace(req.AspectRatio); ar != "" {
			input["aspect_ratio"] = ar
		} else {
			input["aspect_ratio"] = "auto"
		}
		if res := strings.TrimSpace(req.Resolution); res != "" {
			input["resolution"] = res
		} else {
			input["resolution"] = "1k"
		}
		if req.NumImages > 0 {
			input["num_images"] = req.NumImages
		} else {
			input["num_images"] = 1
		}
	}
}
