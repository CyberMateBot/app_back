package ai

import "testing"

func TestBuildKlingCameraControlPreset(t *testing.T) {
	got := buildKlingCameraControl(&CameraControl{Type: "down_back"})
	if got == nil || got["type"] != "down_back" {
		t.Fatalf("preset: %+v", got)
	}
	if _, ok := got["config"]; ok {
		t.Fatalf("preset should not include config: %+v", got)
	}
}

func TestBuildKlingCameraControlSimple(t *testing.T) {
	got := buildKlingCameraControl(&CameraControl{
		Type: "simple",
		Config: &CameraControlConfig{Zoom: 5},
	})
	cfg, ok := got["config"].(map[string]any)
	if !ok || cfg["zoom"] != 5.0 {
		t.Fatalf("simple config: %+v", got)
	}
}

func TestBuildKlingCameraControlAutoSkipped(t *testing.T) {
	if buildKlingCameraControl(&CameraControl{Type: "auto"}) != nil {
		t.Fatal("auto should be omitted")
	}
}

func TestApplyKlingVideoInput(t *testing.T) {
	sound := true
	input := map[string]any{
		"prompt":          "test",
		"duration":        5,
		"generate_audio":  true,
		"camera_fixed":    true,
		"resolution":      "720p",
	}
	req := VideoRequest{
		NegativePrompt: "blur",
		Sound:          &sound,
		CameraControl: &CameraControl{
			Type:   "simple",
			Config: &CameraControlConfig{Pan: -3},
		},
	}

	applyKlingVideoInput(input, mediaModelDef{ID: "kling-v3-pro"}, req)

	if input["negative_prompt"] != "blur" {
		t.Fatalf("negative_prompt: %+v", input)
	}
	if input["sound"] != true {
		t.Fatalf("sound: %+v", input)
	}
	if _, ok := input["generate_audio"]; ok {
		t.Fatal("generate_audio should be removed for kling")
	}
	cc, ok := input["camera_control"].(map[string]any)
	if !ok || cc["type"] != "simple" {
		t.Fatalf("camera_control: %+v", input["camera_control"])
	}
}
