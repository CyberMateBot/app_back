package billing

import "testing"

func TestVideoGenerationPrice_Kling(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("kling-v3-std", "video")
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "kling-v3-std", Duration: 5, Resolution: "720p",
	}); got != 125 {
		t.Fatalf("std 5s 720p: got %d, want 125", got)
	}
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "kling-v3-std", Duration: 5, Resolution: "1080p",
	}); got != 166 {
		t.Fatalf("std 5s 1080p (pro tier): got %d, want 166", got)
	}
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "kling-v3-std", Duration: 5, Resolution: "720p", Sound: true,
	}); got != 155 {
		t.Fatalf("std 5s 720p + sound: got %d, want 155", got)
	}
	base4k := DefaultModelPrice("kling-v3-4k", "video")
	if got := VideoGenerationPrice(base4k, VideoGenerationParams{
		ModelID: "kling-v3-4k", Duration: 5, Resolution: "4k", Sound: false,
	}); got != 250 {
		t.Fatalf("4k 5s: got %d, want 250", got)
	}
	if got := VideoGenerationPrice(base4k, VideoGenerationParams{
		ModelID: "kling-v3-4k", Duration: 5, Resolution: "4k", Sound: true,
	}); got != 280 {
		t.Fatalf("4k 5s + sound: got %d, want 280", got)
	}
	if got := VideoGenerationPrice(base4k, VideoGenerationParams{
		ModelID: "kling-v3-4k", Duration: 10, Resolution: "4k", Sound: false,
	}); got != 500 {
		t.Fatalf("4k 10s: got %d, want 500", got)
	}
	if got := VideoGenerationPrice(base4k, VideoGenerationParams{
		ModelID: "kling-v3-4k", Duration: 10, Resolution: "4k", Sound: true,
	}); got != 560 {
		t.Fatalf("4k 10s + sound: got %d, want 560", got)
	}
}

func TestVideoGenerationPrice_Sora2(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("sora-2-t2v", "video")
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "sora-2-t2v", Duration: 5, Resolution: "720p",
	}); got != 149 {
		t.Fatalf("sora-2 5s 720p: got %d, want 149", got)
	}
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "sora-2-t2v", Duration: 10, Resolution: "720p",
	}); got != 298 {
		t.Fatalf("sora-2 10s 720p: got %d, want 298", got)
	}
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "sora-2-t2v", Duration: 5, Resolution: "1080p",
	}); got != 250 {
		t.Fatalf("sora-2 5s 1080p: got %d, want 250", got)
	}
	basePro := DefaultModelPrice("sora-2-t2v-pro", "video")
	if got := VideoGenerationPrice(basePro, VideoGenerationParams{
		ModelID: "sora-2-t2v-pro", Duration: 5,
	}); got != 250 {
		t.Fatalf("sora-2 pro 5s: got %d, want 250", got)
	}
}

func TestVideoGenerationPrice_Seedance15(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("seedance-v1.5-t2v-fast", "video")
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "seedance-v1.5-t2v-fast", Duration: 5, Resolution: "720p", GenerateAudio: true,
	}); got != 149 {
		t.Fatalf("5s 720p audio on: got %d, want 149", got)
	}
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "seedance-v1.5-t2v-fast", Duration: 5, Resolution: "1080p", GenerateAudio: true,
	}); got != 224 {
		t.Fatalf("5s 1080p audio on: got %d, want 224", got)
	}
}

func TestVideoGenerationPrice_WAN(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("wan-2.5-t2v", "video")
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "wan-2.5-t2v", Duration: 5, Resolution: "720P",
	}); got != 149 {
		t.Fatalf("2.5 5s 720P: got %d, want 149", got)
	}

	base27 := DefaultModelPrice("wan-2.7-t2v", "video")
	if got := VideoGenerationPrice(base27, VideoGenerationParams{
		ModelID: "wan-2.7-t2v", Duration: 5, Resolution: "1080P",
	}); got != 224 {
		t.Fatalf("2.7 5s 1080P: got %d, want 224", got)
	}
}

func TestVideoGenerationPrice_HappyHorse(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("happyhorse-t2v", "video")
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "happyhorse-t2v", Duration: 5, Resolution: "720p",
	}); got != 223 {
		t.Fatalf("5s 720p: got %d, want 223", got)
	}
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "happyhorse-t2v", Duration: 5, Resolution: "1080p",
	}); got != 446 {
		t.Fatalf("5s 1080p: got %d, want 446", got)
	}
}

func TestVideoGenerationPrice_VeoExtend(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("veo-3.1-extend", "video")
	if got := VideoGenerationPrice(base, VideoGenerationParams{ModelID: "veo-3.1-extend"}); got != 312 {
		t.Fatalf("default 4s 1080p audio: got %d, want 312", got)
	}
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "veo-3.1-extend", Duration: 7, Resolution: "1080p", GenerateAudio: true,
	}); got != 546 {
		t.Fatalf("7s 1080p audio: got %d, want 546", got)
	}
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "veo-3.1-extend", Duration: 4, Resolution: "720p", GenerateAudio: true,
	}); got != 265 {
		t.Fatalf("4s 720p audio: got %d, want 265", got)
	}
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "veo-3.1-extend", Duration: 4, Resolution: "1080p", GenerateAudio: false,
	}); got != 209 {
		t.Fatalf("4s 1080p no audio: got %d, want 209", got)
	}
}

func TestVideoGenerationPrice_Vidu(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("vidu-q3-i2v-spicy", "video")
	if got := VideoGenerationPrice(base, VideoGenerationParams{
		ModelID: "vidu-q3-i2v-spicy", Duration: 5, Resolution: "720p",
	}); got != 238 {
		t.Fatalf("5s 720p: got %d, want 238", got)
	}
}

func TestVideoGenerationPrice_Hailuo(t *testing.T) {
	t.Parallel()

	if got := VideoGenerationPrice(0, VideoGenerationParams{ModelID: "hailuo-2.3-t2v", Duration: 6}); got != 44 {
		t.Fatalf("t2v 6s: got %d, want 44", got)
	}
	if got := VideoGenerationPrice(0, VideoGenerationParams{ModelID: "hailuo-2.3-t2v", Duration: 10}); got != 107 {
		t.Fatalf("t2v 10s: got %d, want 107", got)
	}
	if got := VideoGenerationPrice(0, VideoGenerationParams{ModelID: "hailuo-2.3-i2v-fast", Duration: 6}); got != 36 {
		t.Fatalf("fast i2v 6s: got %d, want 36", got)
	}
	if got := VideoGenerationPrice(0, VideoGenerationParams{ModelID: "hailuo-2.3-i2v-fast", Duration: 10}); got != 61 {
		t.Fatalf("fast i2v 10s: got %d, want 61", got)
	}
	if got := VideoGenerationPrice(0, VideoGenerationParams{ModelID: "hailuo-2.3-i2v-pro", Duration: 5}); got != 93 {
		t.Fatalf("pro i2v 5s: got %d, want 93", got)
	}
}
