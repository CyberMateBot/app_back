package ai

import (
	"context"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func generateWavespeedThreeD(ctx context.Context, cfg config.ConfigAI, prompt string, req ThreeDRequest, def mediaModelDef) (ThreeDResponse, error) {
	if err := prepareWavespeedThreeDSource(ctx, cfg, &req); err != nil {
		return ThreeDResponse{}, err
	}

	slug, err := resolveWavespeedThreeDSlug(def, req)
	if err != nil {
		return ThreeDResponse{}, err
	}

	input := buildWavespeedThreeDInput(def, prompt, req)
	urls, err := runWavespeedModel(ctx, cfg, slug, input)
	if err != nil {
		return ThreeDResponse{}, err
	}

	return ThreeDResponse{
		ModelURL: urls[0],
		Model:    def.ID,
	}, nil
}

func resolveWavespeedThreeDSlug(def mediaModelDef, req ThreeDRequest) (string, error) {
	if def.TextSlug == "" {
		return "", &ProviderError{Provider: "wavespeed", Message: "unknown 3d model slug"}
	}

	if def.RequiresMultiImage {
		images := normalizeThreeDImages(req)
		if len(images) < 2 {
			return "", &ProviderError{Provider: "wavespeed", Message: "at least 2 source images are required for multiview 3d"}
		}
		return def.TextSlug, nil
	}

	if def.RequiresImage {
		source := strings.TrimSpace(req.SourceImageURL)
		if source == "" {
			source = strings.TrimSpace(req.ImageURL)
		}
		if source == "" {
			return "", &ProviderError{Provider: "wavespeed", Message: "source image is required for this 3d model"}
		}
	}

	return def.TextSlug, nil
}

func normalizeThreeDImages(req ThreeDRequest) []string {
	if len(req.SourceImages) > 0 {
		return req.SourceImages
	}
	return req.Images
}

func buildWavespeedThreeDInput(def mediaModelDef, prompt string, req ThreeDRequest) map[string]any {
	switch def.ID {
	case "tripo3d-v2.5-i2d":
		return buildTripoV25ImageInput(req)
	case "tripo3d-v2.5-multiview":
		return buildTripoV25MultiviewInput(req)
	case "tripo3d-h3.1-t2d":
		return buildTripoH31TextInput(prompt, req)
	case "tripo3d-h3.1-i2d":
		return buildTripoH31ImageInput(req)
	case "hunyuan3d-v3-t2d":
		return buildHunyuanV3Input(prompt, req)
	case "hunyuan3d-v3.1-rapid":
		return map[string]any{"prompt": strings.TrimSpace(prompt)}
	case "meshy6-t2d":
		return buildMeshy6Input(prompt, req)
	case "rodin-v2-i2d":
		return buildRodinV2Input(prompt, req)
	case "rodin-v2.5-i2d":
		return buildRodinV25Input(prompt, req)
	default:
		return map[string]any{"prompt": strings.TrimSpace(prompt)}
	}
}

func threeDSourceImage(req ThreeDRequest) string {
	source := strings.TrimSpace(req.SourceImageURL)
	if source == "" {
		source = strings.TrimSpace(req.ImageURL)
	}
	return source
}

func buildTripoV25ImageInput(req ThreeDRequest) map[string]any {
	input := map[string]any{
		"image":           threeDSourceImage(req),
		"texture_quality": defaultString(req.TextureQuality, "detailed"),
		"face_limit":      defaultInt(req.FaceLimit, 30000),
		"quad":            req.Quad == nil || *req.Quad,
		"pbr":             req.PBR == nil || *req.PBR,
		"output_format":   defaultString(req.OutputFormat, "glb"),
	}
	return input
}

func buildTripoV25MultiviewInput(req ThreeDRequest) map[string]any {
	return map[string]any{
		"images":          normalizeThreeDImages(req),
		"texture_quality": defaultString(req.TextureQuality, "detailed"),
		"face_limit":      defaultInt(req.FaceLimit, 50000),
		"quad":            req.Quad == nil || *req.Quad,
		"pbr":             req.PBR == nil || *req.PBR,
	}
}

func buildTripoH31TextInput(prompt string, req ThreeDRequest) map[string]any {
	input := map[string]any{
		"prompt":           strings.TrimSpace(prompt),
		"texture":          req.Texture == nil || *req.Texture,
		"pbr":              req.PBR == nil || *req.PBR,
		"texture_quality":  defaultString(req.TextureQuality, "detailed"),
		"geometry_quality": defaultString(req.GeometryQuality, "detailed"),
		"quad":             req.Quad == nil || *req.Quad,
		"auto_size":        req.AutoSize == nil || *req.AutoSize,
		"face_limit":       defaultInt(req.FaceLimit, 50000),
	}
	if neg := strings.TrimSpace(req.NegativePrompt); neg != "" {
		input["negative_prompt"] = neg
	}
	return input
}

func buildTripoH31ImageInput(req ThreeDRequest) map[string]any {
	return map[string]any{
		"image":            threeDSourceImage(req),
		"texture":          req.Texture == nil || *req.Texture,
		"pbr":              req.PBR == nil || *req.PBR,
		"texture_quality":  defaultString(req.TextureQuality, "detailed"),
		"geometry_quality": defaultString(req.GeometryQuality, "detailed"),
		"quad":             req.Quad == nil || *req.Quad,
		"face_limit":       defaultInt(req.FaceLimit, 50000),
	}
}

func buildHunyuanV3Input(prompt string, req ThreeDRequest) map[string]any {
	input := map[string]any{
		"prompt":  strings.TrimSpace(prompt),
		"texture": req.Texture == nil || *req.Texture,
		"pbr":     req.PBR == nil || *req.PBR,
	}
	if neg := strings.TrimSpace(req.NegativePrompt); neg != "" {
		input["negative_prompt"] = neg
	}
	return input
}

func buildMeshy6Input(prompt string, req ThreeDRequest) map[string]any {
	input := map[string]any{
		"prompt":                  strings.TrimSpace(prompt),
		"mode":                    defaultString(req.Mode, "full"),
		"art_style":               defaultString(req.ArtStyle, "realistic"),
		"topology":                defaultString(req.Topology, "quad"),
		"target_polycount":        defaultInt(req.TargetPolycount, 50000),
		"enable_pbr":              req.EnablePBR == nil || *req.EnablePBR,
		"enable_prompt_expansion": req.EnablePromptExpansion == nil || *req.EnablePromptExpansion,
		"ta_pose":                 req.TAPose != nil && *req.TAPose,
		"symmetry_mode":           defaultString(req.SymmetryMode, "auto"),
		"should_remesh":           req.ShouldRemesh == nil || *req.ShouldRemesh,
	}
	return input
}

func buildRodinV2Input(prompt string, req ThreeDRequest) map[string]any {
	input := map[string]any{
		"image":            threeDSourceImage(req),
		"tier":             defaultString(req.Tier, "Gen-2-Medium"),
		"quality_and_mesh": defaultString(req.QualityAndMesh, "8K Quad"),
		"material":         defaultString(req.Material, "PBR"),
		"ta_pose":          req.TAPose != nil && *req.TAPose,
		"hd_texture":       req.HDTexture == nil || *req.HDTexture,
	}
	if p := strings.TrimSpace(prompt); p != "" {
		input["prompt"] = p
	}
	if addons := strings.TrimSpace(req.Addons); addons != "" {
		input["addons"] = addons
	}
	return input
}

func buildRodinV25Input(prompt string, req ThreeDRequest) map[string]any {
	images := normalizeThreeDImages(req)
	source := threeDSourceImage(req)
	if len(images) == 0 && source != "" {
		images = []string{source}
	}
	input := map[string]any{
		"images":                 images,
		"tier":                   defaultString(req.Tier, "Gen-2.5-Medium"),
		"geometry_file_format":   defaultString(req.GeometryFileFormat, "glb"),
		"material":               defaultString(req.Material, "All"),
		"quality_and_mesh":       defaultString(req.QualityAndMesh, "50K Quad"),
		"texture_mode":           defaultString(req.TextureMode, "medium"),
		"geometry_instruct_mode": defaultString(req.GeometryInstructMode, "faithful"),
		"hd_texture":             req.HDTexture == nil || *req.HDTexture,
		"texture_delight":        req.TextureDelight != nil && *req.TextureDelight,
		"is_symmetric":           defaultString(req.IsSymmetric, "auto"),
		"is_micro":               req.IsMicro != nil && *req.IsMicro,
		"ta_pose":                req.TAPose != nil && *req.TAPose,
		"preview_render":         req.PreviewRender == nil || *req.PreviewRender,
	}
	if p := strings.TrimSpace(prompt); p != "" {
		input["prompt"] = p
	}
	if addons := strings.TrimSpace(req.Addons); addons != "" {
		input["addons"] = addons
	}
	return input
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return strings.TrimSpace(value)
}

func defaultInt(value, fallback int) int {
	if value <= 0 {
		return fallback
	}
	return value
}
