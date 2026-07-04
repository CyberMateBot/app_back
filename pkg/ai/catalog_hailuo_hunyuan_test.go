package ai

import "testing"

func TestCatalogHailuoHunyuan(t *testing.T) {
	wantHailuo := map[string]struct {
		requiresImage bool
		durationOpts  int
	}{
		"hailuo-2.3-t2v":      {false, 2},
		"hailuo-2.3-i2v-fast": {true, 2},
		"hailuo-2.3-i2v-pro":  {true, 1},
	}
	found := map[string]bool{}
	for _, m := range ListVideoModels() {
		spec, ok := wantHailuo[m.ID]
		if !ok {
			continue
		}
		found[m.ID] = true
		if m.RequiresImage != spec.requiresImage {
			t.Errorf("%s requires_image: got %v want %v", m.ID, m.RequiresImage, spec.requiresImage)
		}
		var durOpt *MediaOption
		for i := range m.Options {
			if m.Options[i].Key == "duration" {
				durOpt = &m.Options[i]
				break
			}
		}
		if durOpt == nil {
			t.Fatalf("%s: missing duration option", m.ID)
		}
		if len(durOpt.Values) != spec.durationOpts {
			t.Errorf("%s duration values: got %v want %d options", m.ID, durOpt.Values, spec.durationOpts)
		}
		if len(durOpt.ValuePrices) == 0 {
			t.Errorf("%s duration should have value_prices", m.ID)
		}
	}
	for id := range wantHailuo {
		if !found[id] {
			t.Errorf("missing video model %s in catalog", id)
		}
	}

	rapidI2D, ok := findThreeDModel("hunyuan3d-v3.1-rapid-i2d")
	if !ok {
		t.Fatal("hunyuan3d-v3.1-rapid-i2d not in catalog")
	}
	if !rapidI2D.RequiresImage {
		t.Error("hunyuan3d-v3.1-rapid-i2d requires_image should be true")
	}
	hasPBR, hasGeom := false, false
	for _, opt := range rapidI2D.Options {
		switch opt.Key {
		case "enable_pbr":
			hasPBR = true
			if len(opt.ValuePrices) == 0 {
				t.Error("rapid-i2d enable_pbr should have value_prices")
			}
		case "enable_geometry":
			hasGeom = true
		}
	}
	if !hasPBR || !hasGeom {
		t.Fatalf("rapid-i2d options: enable_pbr=%v enable_geometry=%v", hasPBR, hasGeom)
	}

	hunyuanV3, ok := findThreeDModel("hunyuan3d-v3-t2d")
	if !ok {
		t.Fatal("hunyuan3d-v3-t2d not in catalog")
	}
	for _, opt := range hunyuanV3.Options {
		if opt.Key == "generate_type" && len(opt.ValuePrices) == 0 {
			t.Error("hunyuan3d-v3-t2d generate_type should have value_prices")
		}
	}
}

func findThreeDModel(id string) (MediaModel, bool) {
	for _, m := range ListThreeDModels() {
		if m.ID == id {
			return m, true
		}
	}
	return MediaModel{}, false
}
