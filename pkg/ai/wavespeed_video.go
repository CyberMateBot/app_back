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
		if def.ID == "wan-2.7-edit" {
			if instr := strings.TrimSpace(req.EditInstruction); instr != "" {
				input["edit_instruction"] = instr
			} else if strings.TrimSpace(prompt) != "" {
				input["edit_instruction"] = strings.TrimSpace(prompt)
			}
			input["video"] = sourceVideo
		} else {
			input["video"] = sourceVideo
		}
	}

	if def.ID == "happyhorse-video-extend" || def.ID == "veo-3.1-extend" {
		if req.ExtendBy > 0 {
			input["extend_by"] = req.ExtendBy
		}
	}

	if def.ID == "wan-2.7-flf" {
		if first := strings.TrimSpace(req.FirstFrameURL); first != "" {
			input["first_frame_url"] = first
		}
		if last := strings.TrimSpace(req.LastFrameURL); last != "" {
			input["last_frame_url"] = last
		}
	}

	if def.ID == "wan-2.7-grid" {
		gridURL := strings.TrimSpace(req.ImageGridURL)
		if gridURL == "" {
			gridURL = strings.TrimSpace(req.SourceImageURL)
		}
		if gridURL == "" {
			gridURL = strings.TrimSpace(req.ImageURL)
		}
		if gridURL != "" {
			input["image_grid"] = gridURL
		}
	}

	if def.ID == "happyhorse-ref2v" {
		refs := normalizeReferenceImages(req.ReferenceImages)
		if len(refs) == 0 {
			sourceImage := strings.TrimSpace(req.SourceImageURL)
			if sourceImage == "" {
				sourceImage = strings.TrimSpace(req.ImageURL)
			}
			if sourceImage != "" {
				refs = []string{sourceImage}
			}
		}
		if len(refs) > 0 {
			input["reference_images"] = refs
		}
	}

	if def.ID == "vidu-q3-i2v-spicy" {
		if req.BGM != nil {
			input["bgm"] = *req.BGM
		} else {
			input["bgm"] = true
		}
		if amp := strings.TrimSpace(req.MovementAmplitude); amp != "" {
			input["movement_amplitude"] = amp
		} else {
			input["movement_amplitude"] = "auto"
		}
	}

	if res := strings.TrimSpace(req.Resolution); res != "" {
		input["resolution"] = normalizeVideoResolution(def.ID, res)
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
	case strings.HasPrefix(modelID, "wan-"):
		return 5
	case strings.HasPrefix(modelID, "happyhorse-"):
		return 5
	case strings.HasPrefix(modelID, "sora-"):
		return 5
	case strings.HasPrefix(modelID, "hailuo-"):
		return 6
	case strings.HasPrefix(modelID, "vidu-"):
		return 5
	case strings.HasPrefix(modelID, "veo-"):
		return 5
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
	case strings.HasPrefix(modelID, "wan-2.5"), strings.HasPrefix(modelID, "wan-2.6"), strings.HasPrefix(modelID, "wan-2.2"):
		return "720P"
	case strings.HasPrefix(modelID, "wan-2.7"):
		return "1080P"
	case strings.HasPrefix(modelID, "happyhorse-"):
		return "720p"
	case strings.HasPrefix(modelID, "sora-"):
		return "720p"
	case strings.HasPrefix(modelID, "vidu-"):
		return "720p"
	case strings.HasPrefix(modelID, "seedance-v1.5"):
		return "720p"
	case strings.HasPrefix(modelID, "seedance-v2"):
		return "720p"
	default:
		return ""
	}
}

func normalizeVideoResolution(modelID, resolution string) string {
	resolution = strings.TrimSpace(resolution)
	if resolution == "" {
		return resolution
	}
	if strings.HasPrefix(modelID, "wan-") {
		switch strings.ToLower(resolution) {
		case "480p":
			return "480P"
		case "720p":
			return "720P"
		case "1080p":
			return "1080P"
		default:
			return strings.ToUpper(resolution)
		}
	}
	return strings.ToLower(resolution)
}

func modelSupportsGenerateAudio(modelID string) bool {
	return strings.HasPrefix(modelID, "seedance-v1.5") || modelID == "vidu-q3-i2v-spicy"
}

func seedanceUsesCameraFixed(modelID string) bool {
	return strings.HasPrefix(modelID, "seedance-v1")
}

func seedanceUsesSeed(modelID string) bool {
	return strings.HasPrefix(modelID, "seedance-v1")
}
