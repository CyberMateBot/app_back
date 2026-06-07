package ai

import (
	"testing"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func TestExtensionForImageMime(t *testing.T) {
	if extensionForImageMime("image/jpeg") != ".jpg" {
		t.Fatal("expected .jpg")
	}
	if extensionForImageMime("") != ".png" {
		t.Fatal("expected default .png")
	}
}

func TestPrepareWavespeedImageSourceSkipsWhenSourcePresent(t *testing.T) {
	req := ImageRequest{
		SourceImageURL: "https://cdn.example.com/a.png",
		ImageBase64:    "aGVsbG8=",
	}
	if err := prepareWavespeedImageSource(t.Context(), config.ConfigAI{}, &req); err != nil {
		t.Fatal(err)
	}
}
