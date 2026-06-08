package ai

import (
	"context"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func generateWavespeedVideo(ctx context.Context, cfg config.ConfigAI, prompt string, req VideoRequest, def mediaModelDef) (VideoResponse, error) {
	slug, err := resolveWavespeedVideoSlug(def, req)
	if err != nil {
		return VideoResponse{}, err
	}

	enrichedPrompt := enrichMediaPrompt(prompt, req.Messages)
	input := buildWavespeedVideoInput(def, enrichedPrompt, req)
	urls, err := runWavespeedModel(ctx, cfg, slug, input)
	if err != nil {
		return VideoResponse{}, err
	}

	return VideoResponse{
		VideoURL: urls[0],
		Model:    wavespeedVideoEditModelID(def, slug),
	}, nil
}

func buildWavespeedVideoInput(def mediaModelDef, prompt string, req VideoRequest) map[string]any {
	input := map[string]any{}

	if strings.TrimSpace(prompt) != "" {
		input["prompt"] = strings.TrimSpace(prompt)
	}

	duration := req.Duration
	if duration <= 0 {
		duration = defaultVideoDuration(def.ID)
	}
	input["duration"] = duration

	if ar := strings.TrimSpace(req.AspectRatio); ar != "" {
		input["aspect_ratio"] = ar
	} else if !def.RequiresVideo || strings.HasPrefix(def.ID, "seedance-v2") {
		input["aspect_ratio"] = defaultVideoAspectRatio(def.ID)
	}

	if def.RequiresImage {
		sourceImage := strings.TrimSpace(req.SourceImageURL)
		if sourceImage == "" {
			sourceImage = strings.TrimSpace(req.ImageURL)
		}
		input["image"] = sourceImage
	}

	if def.RequiresVideo {
		sourceVideo := strings.TrimSpace(req.SourceVideoURL)
		if sourceVideo == "" {
			sourceVideo = strings.TrimSpace(req.VideoURL)
		}
		input["video"] = sourceVideo
	}

	if res := strings.TrimSpace(req.Resolution); res != "" {
		input["resolution"] = res
	} else if defaultRes := defaultVideoResolution(def.ID); defaultRes != "" {
		input["resolution"] = defaultRes
	}

	if req.GenerateAudio != nil {
		input["generate_audio"] = *req.GenerateAudio
	} else if modelSupportsGenerateAudio(def.ID) {
		input["generate_audio"] = true
	}

	if req.CameraFixed != nil {
		input["camera_fixed"] = *req.CameraFixed
	} else if seedanceUsesCameraFixed(def.ID) {
		input["camera_fixed"] = false
	}

	if req.Seed != 0 {
		input["seed"] = req.Seed
	} else if seedanceUsesSeed(def.ID) {
		input["seed"] = -1
	}

	if lastImage := strings.TrimSpace(req.LastImageURL); lastImage != "" {
		input["last_image"] = lastImage
	}

	if refs := normalizeReferenceImages(req.ReferenceImages); len(refs) > 0 {
		input["reference_images"] = refs
	}

	if isUnifiedSeedanceVideoEdit(def) {
		input["enable_web_search"] = false
	}

	applyKlingVideoInput(input, def, req)

	return input
}

func normalizeReferenceImages(refs []string) []string {
	if len(refs) == 0 {
		return nil
	}
	out := make([]string, 0, len(refs))
	for _, ref := range refs {
		ref = strings.TrimSpace(ref)
		if ref != "" {
			out = append(out, ref)
		}
	}
	return out
}

func defaultVideoDuration(modelID string) int {
	switch {
	case strings.HasPrefix(modelID, "seedance-v1-pro"):
		return 5
	case strings.HasPrefix(modelID, "seedance-v1.5"):
		return 5
	case strings.HasPrefix(modelID, "seedance-v2"):
		return 5
	default:
		return 5
	}
}

func defaultVideoAspectRatio(modelID string) string {
	if strings.HasPrefix(modelID, "seedance-v1-pro") {
		return "16:9"
	}
	return "16:9"
}

func defaultVideoResolution(modelID string) string {
	switch {
	case strings.HasPrefix(modelID, "seedance-v1.5"):
		return "720p"
	case strings.HasPrefix(modelID, "seedance-v2"):
		return "720p"
	default:
		return ""
	}
}

func modelSupportsGenerateAudio(modelID string) bool {
	return strings.HasPrefix(modelID, "seedance-v1.5")
}

func seedanceUsesCameraFixed(modelID string) bool {
	return strings.HasPrefix(modelID, "seedance-v1")
}

func seedanceUsesSeed(modelID string) bool {
	return strings.HasPrefix(modelID, "seedance-v1")
}
