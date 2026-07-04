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
		return buildHunyuanRapidT2DInput(prompt, req)
	case "hunyuan3d-v3.1-rapid-i2d":
		return buildHunyuanRapidI2DInput(req)
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
	return map[string]any{
		"image": threeDSourceImage(req),
	}
}

func buildTripoV25MultiviewInput(req ThreeDRequest) map[string]any {
	images := normalizeThreeDImages(req)
	input := map[string]any{}
	viewKeys := []string{"front_image_url", "back_image_url", "left_image_url", "right_image_url"}
	for i, key := range viewKeys {
		if i < len(images) && strings.TrimSpace(images[i]) != "" {
			input[key] = strings.TrimSpace(images[i])
		}
	}
	return input
}

func buildTripoH31TextInput(prompt string, req ThreeDRequest) map[string]any {
	input := map[string]any{
		"prompt":           strings.TrimSpace(prompt),
		"texture":          req.Texture == nil || *req.Texture,
		"pbr":              req.PBR == nil || *req.PBR,
		"texture_quality":  defaultString(req.TextureQuality, "standard"),
		"geometry_quality": defaultString(req.GeometryQuality, "standard"),
		"quad":             req.Quad != nil && *req.Quad,
		"auto_size":        req.AutoSize != nil && *req.AutoSize,
	}
	if req.FaceLimit > 0 {
		input["face_limit"] = req.FaceLimit
	}
	if neg := strings.TrimSpace(req.NegativePrompt); neg != "" {
		input["negative_prompt"] = neg
	}
	return input
}

func buildTripoH31ImageInput(req ThreeDRequest) map[string]any {
	input := map[string]any{
		"image":            threeDSourceImage(req),
		"texture":          req.Texture == nil || *req.Texture,
		"pbr":              req.PBR == nil || *req.PBR,
		"texture_quality":  defaultString(req.TextureQuality, "standard"),
		"geometry_quality": defaultString(req.GeometryQuality, "standard"),
		"quad":             req.Quad != nil && *req.Quad,
	}
	if req.FaceLimit > 0 {
		input["face_limit"] = req.FaceLimit
	}
	return input
}

func buildHunyuanV3Input(prompt string, req ThreeDRequest) map[string]any {
	enablePBR := false
	if req.EnablePBR != nil {
		enablePBR = *req.EnablePBR
	} else if req.PBR != nil {
		enablePBR = *req.PBR
	}

	input := map[string]any{
		"prompt":        strings.TrimSpace(prompt),
		"generate_type": defaultString(req.GenerateType, "Normal"),
		"enable_pbr":    enablePBR,
		"face_count":    defaultInt(req.FaceLimit, 500000),
		"polygon_type":  "triangle",
	}
	return input
}

func buildHunyuanRapidT2DInput(prompt string, req ThreeDRequest) map[string]any {
	enablePBR := false
	if req.EnablePBR != nil {
		enablePBR = *req.EnablePBR
	} else if req.PBR != nil {
		enablePBR = *req.PBR
	}

	return map[string]any{
		"prompt":        strings.TrimSpace(prompt),
		"generate_type": defaultString(req.GenerateType, "Normal"),
		"enable_pbr":    enablePBR,
		"face_count":    defaultInt(req.FaceLimit, 500000),
	}
}

func buildHunyuanRapidI2DInput(req ThreeDRequest) map[string]any {
	enablePBR := false
	if req.EnablePBR != nil {
		enablePBR = *req.EnablePBR
	} else if req.PBR != nil {
		enablePBR = *req.PBR
	}
	enableGeometry := req.EnableGeometry != nil && *req.EnableGeometry

	return map[string]any{
		"image":            threeDSourceImage(req),
		"enable_pbr":       enablePBR,
		"enable_geometry":  enableGeometry,
	}
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
		"images":           rodinReferenceImages(req),
		"quality_and_mesh": normalizeRodinV2QualityAndMesh(defaultString(req.QualityAndMesh, "8k_Quad")),
		"material":         defaultString(req.Material, "PBR"),
	}
	if p := strings.TrimSpace(prompt); p != "" {
		input["prompt"] = p
	}
	if format := strings.TrimSpace(req.GeometryFileFormat); format != "" {
		input["geometry_file_format"] = format
	}
	if req.TAPose != nil {
		input["ta_pose"] = *req.TAPose
	}
	if req.PreviewRender != nil {
		input["preview_render"] = *req.PreviewRender
	}
	if addons := strings.TrimSpace(req.Addons); addons != "" {
		input["addons"] = addons
	}
	return input
}

func buildRodinV25Input(prompt string, req ThreeDRequest) map[string]any {
	input := map[string]any{
		"images":                 rodinReferenceImages(req),
		"tier":                   defaultString(req.Tier, "Gen-2.5-Medium"),
		"geometry_file_format":   defaultString(req.GeometryFileFormat, "glb"),
		"material":               defaultString(req.Material, "All"),
		"quality_and_mesh":       defaultString(req.QualityAndMesh, "500K Triangle"),
		"geometry_instruct_mode": defaultString(req.GeometryInstructMode, "faithful"),
		"is_symmetric":           normalizeRodinV25Symmetric(defaultString(req.IsSymmetric, "unknown")),
		"ta_pose":                req.TAPose != nil && *req.TAPose,
		"preview_render":         req.PreviewRender != nil && *req.PreviewRender,
	}
	if mode := strings.TrimSpace(req.TextureMode); mode != "" {
		input["texture_mode"] = mode
	}
	if req.HDTexture != nil {
		input["hd_texture"] = *req.HDTexture
	}
	if req.TextureDelight != nil {
		input["texture_delight"] = *req.TextureDelight
	}
	if req.IsMicro != nil {
		input["is_micro"] = *req.IsMicro
	}
	if p := strings.TrimSpace(prompt); p != "" {
		input["prompt"] = p
	}
	if addons := strings.TrimSpace(req.Addons); addons != "" {
		input["addons"] = addons
	}
	return input
}

func rodinReferenceImages(req ThreeDRequest) []string {
	images := normalizeThreeDImages(req)
	source := threeDSourceImage(req)
	if len(images) == 0 && source != "" {
		return []string{source}
	}
	return images
}

func normalizeRodinV2QualityAndMesh(value string) string {
	switch strings.ReplaceAll(strings.TrimSpace(value), " ", "_") {
	case "4K_Quad", "4k_Quad", "4k_quad":
		return "4k_Quad"
	case "8K_Quad", "8k_Quad", "8k_quad":
		return "8k_Quad"
	case "18K_Quad", "18k_Quad", "18k_quad":
		return "18k_Quad"
	case "50K_Quad", "50k_Quad", "50k_quad":
		return "50k_Quad"
	case "2K_Triangle", "2k_triangle":
		return "2K_Triangle"
	case "20K_Triangle", "20k_triangle":
		return "20K_Triangle"
	case "250K_Triangle", "250k_triangle":
		return "250K_Triangle"
	case "500K_Triangle", "500k_triangle":
		return "500K_Triangle"
	default:
		if value == "" {
			return "8k_Quad"
		}
		return value
	}
}

func normalizeRodinV25Symmetric(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "auto", "unknown", "":
		return "unknown"
	case "symmetric", "balanced", "asymmetric":
		return strings.ToLower(strings.TrimSpace(value))
	default:
		return value
	}
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
