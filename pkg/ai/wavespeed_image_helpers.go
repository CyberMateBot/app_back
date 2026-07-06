package ai

import (
	"strconv"
	"strings"
)

// wavespeedImageUsesSingularImage lists models whose text-to-image / img2img endpoint
// accepts a single "image" string (not "images" array). Edit slugs may still require "images".
func wavespeedImageUsesSingularImage(modelID string) bool {
	switch modelID {
	case "qwen-image", "qwen-image-2512", "z-image-base", "z-image-turbo", "flux-dev":
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
		"seedream-v5.0-lite", "flux-dev":
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

func wavespeedImageSlugUsesImagesArray(slug string) bool {
	s := strings.ToLower(strings.TrimSpace(slug))
	if s == "" {
		return false
	}
	if strings.Contains(s, "/edit") || strings.Contains(s, "/edit-") {
		// Same slug for t2i + img2img with singular image field.
		if strings.Contains(s, "z-image/base") {
			return false
		}
		return true
	}
	return strings.Contains(s, "sequential") || strings.Contains(s, "multi")
}

func wavespeedImageSourceUsesPluralField(modelID, slug string) bool {
	if wavespeedImageUsesPluralImages(modelID) || wavespeedImageSlugUsesImagesArray(slug) {
		return true
	}
	return !wavespeedImageUsesSingularImage(modelID)
}

func applyWavespeedImageSourceInput(input map[string]any, modelID, slug, sourceURL string, req ImageRequest) {
	if sourceURL == "" {
		return
	}

	if wavespeedImageSourceUsesPluralField(modelID, slug) {
		input["images"] = []string{sourceURL}
		delete(input, "image")
	} else {
		input["image"] = sourceURL
		delete(input, "images")
	}

	if modelID == "z-image-base" && !strings.HasSuffix(slug, "/edit") && req.Strength > 0 {
		input["strength"] = req.Strength
	}
}

func validateWavespeedImageInput(slug string, input map[string]any) error {
	if !wavespeedImageSlugUsesImagesArray(slug) {
		return nil
	}
	if images, ok := input["images"].([]string); ok && len(images) > 0 && strings.TrimSpace(images[0]) != "" {
		return nil
	}
	return &ProviderError{Provider: "wavespeed", Message: "source image is required for this model"}
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

	if modelID == "nano-banana-2" {
		if req.WebSearch {
			input["enable_web_search"] = true
		}
		if req.ImageSearch {
			input["enable_image_search"] = true
		}
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

	if modelID == "flux-dev" {
		size := wavespeedImageSizeFromRequest(modelID, req)
		if size == "" {
			if req.Width > 0 && req.Height > 0 {
				size = formatPixelSize(req.Width, req.Height)
			} else {
				size = "1024*1024"
			}
		}
		input["size"] = size
		delete(input, "aspect_ratio")

		if req.Seed != 0 {
			input["seed"] = req.Seed
		} else {
			input["seed"] = -1
		}

		if req.Strength > 0 {
			input["strength"] = req.Strength
		} else if strings.TrimSpace(req.SourceImageURL) != "" || strings.TrimSpace(req.ImageURL) != "" {
			input["strength"] = 0.8
		}
	}
}
