package ai

import "testing"

func TestResolveWavespeedImageModel(t *testing.T) {
	def, ok := resolveWavespeedImageModel("flux-dev")
	if !ok || def.Slug != "wavespeed-ai/flux-dev" {
		t.Fatalf("flux-dev: %+v ok=%v", def, ok)
	}

	def, ok = resolveWavespeedImageModel("nano-banana")
	if !ok || def.Slug != "google/nano-banana-pro/text-to-image" {
		t.Fatalf("nano-banana: %+v", def)
	}

	def, ok = resolveWavespeedImageModel("wavespeed-ai/flux-dev")
	if !ok || def.ID != "flux-dev" {
		t.Fatalf("slug alias: %+v", def)
	}
}

func TestResolveWavespeedVideoModel(t *testing.T) {
	def, ok := resolveWavespeedVideoModel("kling-v3-pro")
	if !ok || def.Slug != "kwaivgi/kling-v3.0-pro/text-to-video" {
		t.Fatalf("kling-v3-pro: %+v ok=%v", def, ok)
	}
}

func TestNormalizeWavespeedImageSlug(t *testing.T) {
	got := normalizeWavespeedImageSlug("nano-banana-pro")
	if got != "google/nano-banana-pro/text-to-image" {
		t.Fatalf("got %q", got)
	}
}

func TestListMediaModels(t *testing.T) {
	images := ListImageModels()
	if len(images) < 3 {
		t.Fatalf("expected at least 3 image models, got %d", len(images))
	}
	videos := ListVideoModels()
	if len(videos) < 2 {
		t.Fatalf("expected at least 2 video models, got %d", len(videos))
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
