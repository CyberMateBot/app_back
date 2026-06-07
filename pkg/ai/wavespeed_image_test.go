package ai

import (
	"testing"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func TestBuildWavespeedImageInput_NanoBanana2Text(t *testing.T) {
	def, ok := resolveWavespeedImageModel("nano-banana-2")
	if !ok {
		t.Fatal("model not found")
	}

	cfg := config.ConfigAI{NanoBananaOutputFmt: "jpeg"}
	input := buildWavespeedImageInput(cfg, def, "test prompt", ImageRequest{
		Resolution:   "512px",
		AspectRatio:  "16:9",
		OutputFormat: "webp",
	}, def.TextSlug)

	if input["resolution"] != "0.5k" {
		t.Fatalf("resolution: got %v, want 0.5k", input["resolution"])
	}
	if input["output_format"] != "png" {
		t.Fatalf("output_format: got %v, want png", input["output_format"])
	}
	if input["prompt"] != "test prompt" {
		t.Fatalf("prompt: got %v", input["prompt"])
	}
	if _, ok := input["images"]; ok {
		t.Fatalf("text-to-image must not set images")
	}
}

func TestBuildWavespeedImageInput_NanoBanana2Edit(t *testing.T) {
	def, ok := resolveWavespeedImageModel("nano-banana-2")
	if !ok {
		t.Fatal("model not found")
	}

	cfg := config.ConfigAI{}
	input := buildWavespeedImageInput(cfg, def, "edit prompt", ImageRequest{
		SourceImageURL: "https://cdn.example.com/source.png",
		Resolution:     "2k",
		AspectRatio:    "1:1",
		OutputFormat:   "jpeg",
	}, def.EditSlug)

	images, ok := input["images"].([]string)
	if !ok || len(images) != 1 || images[0] != "https://cdn.example.com/source.png" {
		t.Fatalf("images: got %v", input["images"])
	}
	if _, ok := input["image"]; ok {
		t.Fatalf("edit must use images, not image")
	}
	if input["resolution"] != "2k" {
		t.Fatalf("resolution: got %v, want 2k", input["resolution"])
	}
}

func TestNormalizeImageResolution(t *testing.T) {
	cases := map[string]string{
		"512px": "0.5k",
		"512":   "0.5k",
		"0.5K":  "0.5k",
		"1K":    "1k",
		"4k":    "4k",
	}
	for in, want := range cases {
		got := normalizeImageResolution("nano-banana-2", in)
		if got != want {
			t.Fatalf("%q: got %q, want %q", in, got, want)
		}
	}
}
