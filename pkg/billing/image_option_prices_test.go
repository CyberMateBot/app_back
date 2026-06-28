package billing

import "testing"

func TestImageGenerationPrice_NanoBanana2(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("nano-banana-2", "image")
	cases := map[string]int{
		"0.5k": 14,
		"1k":   22,
		"2k":   33,
		"4k":   44,
	}
	for res, want := range cases {
		got := ImageGenerationPrice(base, ImageGenerationParams{
			ModelID:    "nano-banana-2",
			Resolution: res,
		})
		if got != want {
			t.Fatalf("resolution %q: got %d, want %d", res, got, want)
		}
	}

	withSearch := ImageGenerationPrice(base, ImageGenerationParams{
		ModelID:     "nano-banana-2",
		Resolution:  "1k",
		WebSearch:   true,
		ImageSearch: true,
	})
	if withSearch != 31 {
		t.Fatalf("1k + searches: got %d, want 31", withSearch)
	}
}

func TestImageGenerationPrice_NanoBananaPro(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("nano-banana-pro", "image")
	if got := ImageGenerationPrice(base, ImageGenerationParams{ModelID: "nano-banana-pro", Resolution: "1k"}); got != 40 {
		t.Fatalf("1k: got %d, want 40", got)
	}
	if got := ImageGenerationPrice(base, ImageGenerationParams{ModelID: "nano-banana-pro", Resolution: "2k"}); got != 40 {
		t.Fatalf("2k: got %d, want 40", got)
	}
	if got := ImageGenerationPrice(base, ImageGenerationParams{ModelID: "nano-banana-pro", Resolution: "4k"}); got != 69 {
		t.Fatalf("4k: got %d, want 69", got)
	}
}

func TestImageGenerationPrice_GPTImage2(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("gpt-image-2", "image")
	if got := ImageGenerationPrice(base, ImageGenerationParams{ModelID: "gpt-image-2"}); got != 20 {
		t.Fatalf("default: got %d, want 20", got)
	}
	if got := ImageGenerationPrice(base, ImageGenerationParams{
		ModelID: "gpt-image-2", Quality: "high", Resolution: "4k",
	}); got != 220 {
		t.Fatalf("high 4k: got %d, want 220", got)
	}
}

func TestImageGenerationPrice_GrokImagineEdit(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("grok-imagine-edit", "image")
	if got := ImageGenerationPrice(base, ImageGenerationParams{ModelID: "grok-imagine-edit"}); got != 25 {
		t.Fatalf("default: got %d, want 25", got)
	}
	if got := ImageGenerationPrice(base, ImageGenerationParams{
		ModelID: "grok-imagine-edit", Resolution: "2k", NumImages: 2,
	}); got != 64 {
		t.Fatalf("2k x2: got %d, want 64", got)
	}
}

func TestImageGenerationPrice_QwenImageSize(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("qwen-image", "image")
	if got := ImageGenerationPrice(base, ImageGenerationParams{ModelID: "qwen-image", Size: "1024*1024"}); got != 7 {
		t.Fatalf("1024: got %d, want 7", got)
	}
	large := ImageGenerationPrice(base, ImageGenerationParams{ModelID: "qwen-image", Size: "1328*1328"})
	if large <= 7 {
		t.Fatalf("1328 should cost more than base: got %d", large)
	}
}

func TestImageGenerationPrice_FlatModels(t *testing.T) {
	t.Parallel()

	for _, modelID := range []string{"flux-dev", "seedream-v4.5", "qwen-image-2.0", "alice-ai-art"} {
		base := DefaultModelPrice(modelID, "image")
		if got := ImageGenerationPrice(base, ImageGenerationParams{
			ModelID: modelID, Resolution: "4k", Quality: "high", Size: "1328*1328",
		}); got != base {
			t.Fatalf("%s should stay flat: got %d, want %d", modelID, got, base)
		}
	}
}

func TestImageOptionValuePrices_NanoBanana2Resolution(t *testing.T) {
	t.Parallel()

	deltas := ImageOptionValuePrices("nano-banana-2", "resolution")
	want := map[string]int{"0.5k": -8, "1k": 0, "2k": 11, "4k": 22}
	for k, v := range want {
		if deltas[k] != v {
			t.Fatalf("delta %q: got %d, want %d", k, deltas[k], v)
		}
	}
}

func TestImageGenerationPrice_NanoBananaFlat(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("nano-banana", "image")
	if got := ImageGenerationPrice(base, ImageGenerationParams{ModelID: "nano-banana", Resolution: "4k"}); got != base {
		t.Fatalf("nano-banana should ignore resolution: got %d, want %d", got, base)
	}
}
