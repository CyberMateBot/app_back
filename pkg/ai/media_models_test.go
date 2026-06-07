package ai

import "testing"

func TestResolveWavespeedImageModel(t *testing.T) {
	def, ok := resolveWavespeedImageModel("flux-dev")
	if !ok || def.TextSlug != "wavespeed-ai/flux-dev" {
		t.Fatalf("flux-dev: %+v ok=%v", def, ok)
	}

	def, ok = resolveWavespeedImageModel("nano-banana")
	if !ok || def.TextSlug != "google/nano-banana/text-to-image" || def.EditSlug != "google/nano-banana/edit" {
		t.Fatalf("nano-banana: %+v", def)
	}

	def, ok = resolveWavespeedImageModel("nano-banana-pro")
	if !ok || def.MultiSlug != "google/nano-banana-pro/text-to-image-multi" {
		t.Fatalf("nano-banana-pro: %+v", def)
	}

	def, ok = resolveWavespeedImageModel("nano-banana-2")
	if !ok || def.TextSlug != "google/nano-banana-2/text-to-image" {
		t.Fatalf("nano-banana-2: %+v", def)
	}

	def, ok = resolveWavespeedImageModel("gpt-image-2")
	if !ok || def.TextSlug != "openai/gpt-image-2/text-to-image" || def.EditSlug != "openai/gpt-image-2/edit" {
		t.Fatalf("gpt-image-2: %+v", def)
	}

	def, ok = resolveWavespeedImageModel("gpt-image-1.5")
	if !ok || def.EditSlug != "openai/gpt-image-1.5/edit" {
		t.Fatalf("gpt-image-1.5: %+v", def)
	}
}

func TestResolveWavespeedImageSlug(t *testing.T) {
	def, _ := resolveWavespeedImageModel("nano-banana")

	slug, err := resolveWavespeedImageSlug(def, ImageRequest{Prompt: "cat"})
	if err != nil || slug != "google/nano-banana/text-to-image" {
		t.Fatalf("text: slug=%q err=%v", slug, err)
	}

	slug, err = resolveWavespeedImageSlug(def, ImageRequest{
		Prompt:         "night scene",
		SourceImageURL: "https://cdn.example.com/a.png",
	})
	if err != nil || slug != "google/nano-banana/edit" {
		t.Fatalf("edit: slug=%q err=%v", slug, err)
	}

	pro, _ := resolveWavespeedImageModel("nano-banana-pro")
	slug, err = resolveWavespeedImageSlug(pro, ImageRequest{Prompt: "landscape", Mode: "multi"})
	if err != nil || slug != "google/nano-banana-pro/text-to-image-multi" {
		t.Fatalf("multi: slug=%q err=%v", slug, err)
	}
}

func TestResolveWavespeedVideoModel(t *testing.T) {
	def, ok := resolveWavespeedVideoModel("kling-v3-pro")
	if !ok || def.TextSlug != "kwaivgi/kling-v3.0-pro/text-to-video" {
		t.Fatalf("kling-v3-pro: %+v ok=%v", def, ok)
	}

	def, ok = resolveWavespeedVideoModel("seedance-v1.5-t2v-fast")
	if !ok || def.TextSlug != "bytedance/seedance-v1.5-pro/text-to-video-fast" {
		t.Fatalf("seedance-v1.5-t2v-fast: %+v", def)
	}

	def, ok = resolveWavespeedVideoModel("seedance-v2-video-edit")
	if !ok || !def.RequiresVideo {
		t.Fatalf("seedance-v2-video-edit: %+v", def)
	}
}

func TestResolveWavespeedVideoSlug(t *testing.T) {
	i2v, _ := resolveWavespeedVideoModel("seedance-v1-pro-i2v")
	_, err := resolveWavespeedVideoSlug(i2v, VideoRequest{Prompt: "sunset"})
	if err == nil {
		t.Fatal("expected error without source image")
	}

	slug, err := resolveWavespeedVideoSlug(i2v, VideoRequest{
		Prompt:         "sunset",
		SourceImageURL: "https://cdn.example.com/a.jpg",
	})
	if err != nil || slug != "bytedance/seedance-v1-pro-i2v-720p" {
		t.Fatalf("i2v slug=%q err=%v", slug, err)
	}

	edit, _ := resolveWavespeedVideoModel("seedance-v2-video-edit")
	slug, err = resolveWavespeedVideoSlug(edit, VideoRequest{
		Prompt:         "cyberpunk night",
		SourceVideoURL: "https://cdn.example.com/v.mp4",
		Resolution:     "480p",
	})
	if err != nil || slug != "bytedance/seedance-2.0/video-edit" {
		t.Fatalf("edit 480p slug=%q err=%v", slug, err)
	}

	slug, err = resolveWavespeedVideoSlug(edit, VideoRequest{
		Prompt:         "cyberpunk night",
		SourceVideoURL: "https://cdn.example.com/v.mp4",
		Resolution:     "1080p",
	})
	if err != nil || slug != "bytedance/seedance-2.0/video-edit-turbo" {
		t.Fatalf("edit 1080p slug=%q err=%v", slug, err)
	}
}

func TestListMediaModels(t *testing.T) {
	images := ListImageModels()
	videos := ListVideoModels()
	if len(videos) < 8 {
		t.Fatalf("expected at least 8 video models, got %d", len(videos))
	}

	if len(images) < 7 {
		t.Fatalf("expected at least 7 image models, got %d", len(images))
	}
	nanoCount := 0
	for _, m := range images {
		if m.Group == "Nano Banana" {
			nanoCount++
		}
		if m.ID == "nano-banana" && !m.SupportsEdit {
			t.Fatalf("nano-banana should support edit")
		}
	}
	if nanoCount < 3 {
		t.Fatalf("expected 3 nano banana models, got %d", nanoCount)
	}

	gptCount := 0
	for _, m := range images {
		if m.Group == "GPT Image" {
			gptCount++
		}
		if m.ID == "gpt-image-2" && len(m.Options) == 0 {
			t.Fatalf("gpt-image-2 should expose options")
		}
	}
	if gptCount < 2 {
		t.Fatalf("expected 2 GPT Image models, got %d", gptCount)
	}
}

func TestExtractWavespeedOutputs(t *testing.T) {
	urls, err := extractWavespeedOutputs(map[string]any{
		"outputs": []any{"https://example.com/a.png"},
	})
	if err != nil || len(urls) != 1 || urls[0] != "https://example.com/a.png" {
		t.Fatalf("extract: urls=%v err=%v", urls, err)
	}
}
