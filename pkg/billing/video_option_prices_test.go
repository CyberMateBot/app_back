package billing

import "testing"

func TestVideoGenerationPrice_Kling(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("kling-v3-std", "video")
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "kling-v3-std", Duration: 5, Resolution: "720p",
	}); got != 80 {
		t.Fatalf("std 5s 720p: got %d, want 80", got)
	}
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "kling-v3-std", Duration: 5, Resolution: "1080p",
	}); got != 110 {
		t.Fatalf("std 5s 1080p (pro tier): got %d, want 110", got)
	}
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "kling-v3-std", Duration: 5, Resolution: "720p", Sound: true,
	}); got != 120 {
		t.Fatalf("std 5s 720p + sound: got %d, want 120", got)
	}
}

func TestVideoGenerationPrice_Seedance15(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("seedance-v1.5-t2v-fast", "video")
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "seedance-v1.5-t2v-fast", Duration: 5, Resolution: "720p", GenerateAudio: true,
	}); got != 110 {
		t.Fatalf("5s 720p audio on: got %d, want 110", got)
	}
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "seedance-v1.5-t2v-fast", Duration: 5, Resolution: "1080p", GenerateAudio: true,
	}); got != 165 {
		t.Fatalf("5s 1080p audio on: got %d, want 165", got)
	}
}

func TestVideoGenerationPrice_WAN(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("wan-2.5-t2v", "video")
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "wan-2.5-t2v", Duration: 5, Resolution: "720P",
	}); got != 95 {
		t.Fatalf("2.5 5s 720P: got %d, want 95", got)
	}

	base27 := DefaultModelPrice("wan-2.7-t2v", "video")
	if got := VideoGenerationPrice(base27, VideoGenerationParams{
		ModelID: "wan-2.7-t2v", Duration: 5, Resolution: "1080P",
	}); got != 143 {
		t.Fatalf("2.7 5s 1080P: got %d, want 143", got)
	}
}

func TestVideoGenerationPrice_HappyHorse(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("happyhorse-t2v", "video")
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "happyhorse-t2v", Duration: 5, Resolution: "720p",
	}); got != 133 {
		t.Fatalf("5s 720p: got %d, want 133", got)
	}
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "happyhorse-t2v", Duration: 5, Resolution: "1080p",
	}); got != 266 {
		t.Fatalf("5s 1080p: got %d, want 266", got)
	}
}

func TestVideoGenerationPrice_VeoExtend(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("veo-3.1-extend", "video")
	if got := VideoGenerationPrice(base, VideoGenerationParams{ModelID: "veo-3.1-extend"}); got != 105 {
		t.Fatalf("flat extend: got %d, want 105", got)
	}
}

func TestVideoGenerationPrice_Vidu(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("vidu-q3-i2v-spicy", "video")
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "vidu-q3-i2v-spicy", Duration: 5, Resolution: "720p",
	}); got != 143 {
		t.Fatalf("5s 720p: got %d, want 143", got)
	}
}

func TestVideoGenerationPrice_Hailuo(t *testing.T) {
	t.Parallel()

	if got := VideoGenerationPrice(0, VideoGenerationParams{ModelID: "hailuo-2.3-t2v", Duration: 6}); got != 44 {
		t.Fatalf("t2v 6s: got %d, want 44", got)
	}
	if got := VideoGenerationPrice(0, VideoGenerationParams{ModelID: "hailuo-2.3-i2v-fast", Duration: 6}); got != 36 {
		t.Fatalf("fast i2v: got %d, want 36", got)
	}
	if got := VideoGenerationPrice(0, VideoGenerationParams{ModelID: "hailuo-2.3-i2v-pro", Duration: 6}); got != 93 {
		t.Fatalf("pro i2v: got %d, want 93", got)
	}
}
