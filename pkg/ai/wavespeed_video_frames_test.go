package ai

import "testing"

func TestBuildWavespeedVideoInputFrames(t *testing.T) {
	req := VideoRequest{
		FirstFrameURL: "https://example.com/first.png",
		LastFrameURL:  "https://example.com/last.png",
	}

	// 1. WAN 2.7 FLF maps to first_frame_url and last_frame_url
	wanDef := mediaModelDef{ID: "wan-2.7-flf"}
	wanInput := buildWavespeedVideoInput(wanDef, "test prompt", req)
	if wanInput["first_frame_url"] != "https://example.com/first.png" {
		t.Fatalf("wan first_frame_url got %v", wanInput["first_frame_url"])
	}
	if wanInput["last_frame_url"] != "https://example.com/last.png" {
		t.Fatalf("wan last_frame_url got %v", wanInput["last_frame_url"])
	}

	// 2. Kling maps to image and image_tail
	klingDef := mediaModelDef{ID: "kling-v3-std"}
	klingInput := buildWavespeedVideoInput(klingDef, "test prompt", req)
	if klingInput["image"] != "https://example.com/first.png" {
		t.Fatalf("kling image got %v", klingInput["image"])
	}
	if klingInput["image_tail"] != "https://example.com/last.png" {
		t.Fatalf("kling image_tail got %v", klingInput["image_tail"])
	}

	// 3. Fallback sourceImageUrl / lastImage works interchangeably
	reqAlt := VideoRequest{
		SourceImageURL: "https://example.com/start.png",
		LastImageURL:   "https://example.com/end.png",
	}
	klingInputAlt := buildWavespeedVideoInput(klingDef, "test prompt", reqAlt)
	if klingInputAlt["image"] != "https://example.com/start.png" {
		t.Fatalf("kling alt image got %v", klingInputAlt["image"])
	}
	if klingInputAlt["image_tail"] != "https://example.com/end.png" {
		t.Fatalf("kling alt image_tail got %v", klingInputAlt["image_tail"])
	}
}

func TestCatalogSupportsLastFrame(t *testing.T) {
	for _, m := range ListVideoModels() {
		if m.ID == "wan-2.7-flf" || isKlingVideoModel(m.ID) {
			if !m.SupportsLastFrame {
				t.Fatalf("model %s should have SupportsLastFrame=true", m.ID)
			}
		}
	}
}
