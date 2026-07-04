package billing

import "testing"

func TestThreeDGenerationPrice_TripoH31(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("tripo3d-h3.1-i2d", "3d")
	if got := ThreeDGenerationPrice(base, ThreeDGenerationParams{
		ModelID: "tripo3d-h3.1-i2d", Texture: true, TextureSet: true, TextureQuality: "standard", GeometryQuality: "standard",
	}); got != 55 {
		t.Fatalf("default: got %d, want 55", got)
	}
	if got := ThreeDGenerationPrice(base, ThreeDGenerationParams{
		ModelID: "tripo3d-h3.1-i2d", Texture: true, TextureSet: true, TextureQuality: "detailed", GeometryQuality: "detailed", Quad: true,
	}); got != 119 {
		t.Fatalf("detailed all + quad: got %d, want 119", got)
	}
	if got := ThreeDGenerationPrice(base, ThreeDGenerationParams{
		ModelID: "tripo3d-h3.1-i2d", Texture: false, TextureSet: true, GeometryQuality: "standard",
	}); got != 37 {
		t.Fatalf("no texture: got %d, want 37", got)
	}
}

func TestThreeDGenerationPrice_HunyuanV3(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("hunyuan3d-v3-t2d", "3d")
	if got := ThreeDGenerationPrice(base, ThreeDGenerationParams{ModelID: "hunyuan3d-v3-t2d", GenerateType: "Normal"}); got != 30 {
		t.Fatalf("normal: got %d, want 30", got)
	}
	if got := ThreeDGenerationPrice(base, ThreeDGenerationParams{ModelID: "hunyuan3d-v3-t2d", GenerateType: "Geometry"}); got != 20 {
		t.Fatalf("geometry: got %d, want 20", got)
	}
	if got := ThreeDGenerationPrice(base, ThreeDGenerationParams{ModelID: "hunyuan3d-v3-t2d", GenerateType: "LowPoly"}); got != 36 {
		t.Fatalf("lowpoly: got %d, want 36", got)
	}
}

func TestThreeDGenerationPrice_HunyuanRapid(t *testing.T) {
	t.Parallel()

	baseT2D := DefaultModelPrice("hunyuan3d-v3.1-rapid", "3d")
	if got := ThreeDGenerationPrice(baseT2D, ThreeDGenerationParams{ModelID: "hunyuan3d-v3.1-rapid"}); got != 25 {
		t.Fatalf("rapid t2d base: got %d, want 25", got)
	}
	if got := ThreeDGenerationPrice(baseT2D, ThreeDGenerationParams{ModelID: "hunyuan3d-v3.1-rapid", EnablePBR: true, EnablePBRSet: true}); got != 192 {
		t.Fatalf("rapid t2d pbr: got %d, want 192", got)
	}

	baseI2D := DefaultModelPrice("hunyuan3d-v3.1-rapid-i2d", "3d")
	if got := ThreeDGenerationPrice(baseI2D, ThreeDGenerationParams{ModelID: "hunyuan3d-v3.1-rapid-i2d"}); got != 250 {
		t.Fatalf("rapid i2d base: got %d, want 250", got)
	}
	if got := ThreeDGenerationPrice(baseI2D, ThreeDGenerationParams{ModelID: "hunyuan3d-v3.1-rapid-i2d", EnablePBR: true, EnablePBRSet: true}); got != 417 {
		t.Fatalf("rapid i2d pbr: got %d, want 417", got)
	}
}

func TestThreeDGenerationPrice_FlatModels(t *testing.T) {
	t.Parallel()

	cases := map[string]int{
		"tripo3d-v2.5-i2d": 48,
		"meshy6-t2d":       48,
		"rodin-v2-i2d":     55,
	}
	for modelID, want := range cases {
		base := DefaultModelPrice(modelID, "3d")
		if got := ThreeDGenerationPrice(base, ThreeDGenerationParams{ModelID: modelID}); got != want {
			t.Fatalf("%s: got %d, want %d", modelID, got, want)
		}
	}
}

func TestThreeDGenerationPrice_RodinV25(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("rodin-v2.5-i2d", "3d")
	if got := ThreeDGenerationPrice(base, ThreeDGenerationParams{ModelID: "rodin-v2.5-i2d", Tier: "Gen-2.5-Medium"}); got != 55 {
		t.Fatalf("medium: got %d, want 55", got)
	}
	if got := ThreeDGenerationPrice(base, ThreeDGenerationParams{ModelID: "rodin-v2.5-i2d", Tier: "Gen-2.5-Extreme-High"}); got != 110 {
		t.Fatalf("extreme high: got %d, want 110", got)
	}
	if got := ThreeDGenerationPrice(base, ThreeDGenerationParams{ModelID: "rodin-v2.5-i2d", Tier: "Gen-2.5-Medium", Addons: "HighPack"}); got != 165 {
		t.Fatalf("highpack: got %d, want 165", got)
	}
}
