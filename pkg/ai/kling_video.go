package ai

import "strings"

const (
	klingCameraTypeAuto              = "auto"
	klingCameraTypeSimple            = "simple"
	klingCameraTypeDownBack          = "down_back"
	klingCameraTypeForwardUp         = "forward_up"
	klingCameraTypeRightTurnForward  = "right_turn_forward"
	klingCameraTypeLeftTurnForward   = "left_turn_forward"
)

func isKlingVideoModel(modelID string) bool {
	return strings.HasPrefix(strings.TrimSpace(modelID), "kling-")
}

func klingDurationOptionValues() []string {
	out := make([]string, 0, 13)
	for sec := 3; sec <= 15; sec++ {
		out = append(out, itoa(sec))
	}
	return out
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var digits []byte
	for n > 0 {
		digits = append([]byte{byte('0' + n%10)}, digits...)
		n /= 10
	}
	return string(digits)
}

func applyKlingVideoInput(input map[string]any, def mediaModelDef, req VideoRequest) {
	if !isKlingVideoModel(def.ID) {
		return
	}

	delete(input, "generate_audio")
	delete(input, "camera_fixed")
	delete(input, "resolution")
	delete(input, "turbo_mode")

	if np := strings.TrimSpace(req.NegativePrompt); np != "" {
		input["negative_prompt"] = np
	}

	switch {
	case req.Sound != nil:
		input["sound"] = *req.Sound
	case req.GenerateAudio != nil && klingModelSupportsSound(def.ID):
		input["sound"] = *req.GenerateAudio
	case def.ID == "kling-v3-pro", def.ID == "kling-v3-4k":
		input["sound"] = false
	}

	if cc := buildKlingCameraControl(req.CameraControl); cc != nil {
		input["camera_control"] = cc
	}
}

func buildKlingCameraControl(cc *CameraControl) map[string]any {
	if cc == nil {
		return nil
	}

	typ := strings.TrimSpace(cc.Type)
	if typ == "" || typ == klingCameraTypeAuto {
		return nil
	}

	out := map[string]any{"type": typ}
	if typ != klingCameraTypeSimple {
		return out
	}

	cfg := buildKlingCameraConfig(cc.Config)
	if len(cfg) == 0 {
		return nil
	}

	out["config"] = cfg
	return out
}

func buildKlingCameraConfig(cfg *CameraControlConfig) map[string]any {
	if cfg == nil {
		return nil
	}

	axes := []struct {
		key string
		val float64
	}{
		{"horizontal", cfg.Horizontal},
		{"vertical", cfg.Vertical},
		{"pan", cfg.Pan},
		{"tilt", cfg.Tilt},
		{"roll", cfg.Roll},
		{"zoom", cfg.Zoom},
	}

	out := map[string]any{}
	nonZero := 0
	for _, axis := range axes {
		if axis.val != 0 {
			nonZero++
			out[axis.key] = clampKlingAxis(axis.val)
		}
	}

	if nonZero == 0 {
		return nil
	}

	// Kling simple mode: only one axis may be non-zero; keep the last non-zero set.
	if nonZero > 1 {
		for _, axis := range axes {
			if axis.val != 0 {
				return map[string]any{axis.key: clampKlingAxis(axis.val)}
			}
		}
	}

	return out
}

func clampKlingAxis(v float64) float64 {
	if v < -10 {
		return -10
	}
	if v > 10 {
		return 10
	}
	return v
}

func klingModelSupportsSound(modelID string) bool {
	return modelID == "kling-v3-pro" || modelID == "kling-v3-4k"
}

func resolveKlingModelID(resolution, fallbackModel string) string {
	switch strings.ToLower(strings.TrimSpace(resolution)) {
	case "4k":
		return "kling-v3-4k"
	case "1080p":
		return "kling-v3-pro"
	case "720p":
		return "kling-v3-std"
	default:
		if isKlingVideoModel(fallbackModel) {
			return fallbackModel
		}
		return "kling-v3-std"
	}
}
