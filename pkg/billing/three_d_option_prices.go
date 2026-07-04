package billing

import "strings"

// ThreeDGenerationParams holds 3D option values that affect generation cost.
type ThreeDGenerationParams struct {
	ModelID         string
	Texture         bool
	TextureSet      bool
	TextureQuality  string
	GeometryQuality string
	Quad            bool
	GenerateType    string
	EnablePBR       bool
	EnablePBRSet    bool
	Tier            string
	Addons          string
}

const (
	tripoV25USD          = 0.30
	tripoH31BaseUSD      = 0.20
	tripoH31TextureUSD   = 0.10
	tripoH31DetailedTex  = 0.20
	tripoH31DetailedGeom = 0.20
	tripoH31QuadUSD      = 0.05

	hunyuanRapidUSD    = 0.0225
	hunyuanRapidI2DBaseUSD = 0.225
	hunyuanRapidPBRUSD = 0.15
	hunyuanGeometryUSD = 0.25
	hunyuanNormalUSD   = 0.375
	hunyuanLowPolyUSD  = 0.45

	meshy6USD        = 0.80
	rodinV2USD       = 0.30
	rodinV25BaseUSD  = 0.40
	rodinHighPackUSD = 0.80
)

// ThreeDOptionValuePrices returns coin surcharges per option value relative to the model base price.
func ThreeDOptionValuePrices(modelID, optionKey string) map[string]int {
	modelID = strings.ToLower(strings.TrimSpace(modelID))
	optionKey = strings.ToLower(strings.TrimSpace(optionKey))
	base := DefaultModelPrice(modelID, "3d")

	switch modelID {
	case "tripo3d-h3.1-t2d", "tripo3d-h3.1-i2d":
		switch optionKey {
		case "texture":
			return tripoH31TextureDeltas(modelID, base)
		case "texture_quality":
			return tripoH31TextureQualityDeltas(modelID, base)
		case "geometry_quality":
			return tripoH31GeometryQualityDeltas(modelID, base)
		case "quad":
			return tripoH31QuadDeltas(modelID, base)
		}
	case "hunyuan3d-v3-t2d":
		if optionKey == "generate_type" {
			return hunyuanGenerateTypeDeltas(modelID, base)
		}
	case "hunyuan3d-v3.1-rapid", "hunyuan3d-v3.1-rapid-i2d":
		if optionKey == "enable_pbr" {
			return hunyuanRapidPBRDeltas(modelID, base)
		}
	case "rodin-v2.5-i2d":
		switch optionKey {
		case "tier":
			return rodinV25TierDeltas(modelID, base)
		case "addons":
			return rodinV25AddonsDeltas(modelID, base)
		}
	}
	return nil
}

// ThreeDGenerationPrice returns total CyberCoins for a 3D generation with the given options.
func ThreeDGenerationPrice(basePrice int, p ThreeDGenerationParams) int {
	modelID := strings.ToLower(strings.TrimSpace(p.ModelID))
	if basePrice <= 0 {
		basePrice = DefaultModelPrice(modelID, "3d")
	}

	usd := threeDGenerationUSD(p)
	if usd <= 0 {
		return basePrice
	}

	baseUSD := defaultThreeDUSD(modelID)
	price := usdToCoinsFromBase(modelID, basePrice, baseUSD, usd)
	if price < 1 {
		return 1
	}
	return price
}

func threeDGenerationUSD(p ThreeDGenerationParams) float64 {
	modelID := strings.ToLower(strings.TrimSpace(p.ModelID))
	switch modelID {
	case "tripo3d-v2.5-i2d", "tripo3d-v2.5-multiview":
		return tripoV25USD
	case "tripo3d-h3.1-t2d", "tripo3d-h3.1-i2d":
		return tripoH31USD(p)
	case "hunyuan3d-v3.1-rapid":
		return hunyuanRapidT2DUSD(p)
	case "hunyuan3d-v3.1-rapid-i2d":
		return hunyuanRapidI2DUSD(p)
	case "hunyuan3d-v3-t2d":
		return hunyuanV3USD(p.GenerateType)
	case "meshy6-t2d":
		return meshy6USD
	case "rodin-v2-i2d":
		return rodinV2USD
	case "rodin-v2.5-i2d":
		return rodinV25USD(p.Tier, p.Addons)
	default:
		return 0
	}
}

func defaultThreeDUSD(modelID string) float64 {
	switch strings.ToLower(strings.TrimSpace(modelID)) {
	case "tripo3d-v2.5-i2d", "tripo3d-v2.5-multiview":
		return tripoV25USD
	case "tripo3d-h3.1-t2d", "tripo3d-h3.1-i2d":
		return tripoH31USD(ThreeDGenerationParams{ModelID: modelID, Texture: true, TextureSet: true, TextureQuality: "standard", GeometryQuality: "standard"})
	case "hunyuan3d-v3.1-rapid":
		return hunyuanRapidUSD
	case "hunyuan3d-v3.1-rapid-i2d":
		return hunyuanRapidI2DBaseUSD
	case "hunyuan3d-v3-t2d":
		return hunyuanNormalUSD
	case "meshy6-t2d":
		return meshy6USD
	case "rodin-v2-i2d":
		return rodinV2USD
	case "rodin-v2.5-i2d":
		return rodinV25BaseUSD
	default:
		return 0
	}
}

func tripoH31USD(p ThreeDGenerationParams) float64 {
	texture := true
	if p.TextureSet {
		texture = p.Texture
	}

	usd := tripoH31BaseUSD
	if texture {
		if strings.EqualFold(strings.TrimSpace(p.TextureQuality), "detailed") {
			usd += tripoH31DetailedTex
		} else {
			usd += tripoH31TextureUSD
		}
	}
	if strings.EqualFold(strings.TrimSpace(p.GeometryQuality), "detailed") {
		usd += tripoH31DetailedGeom
	}
	if p.Quad {
		usd += tripoH31QuadUSD
	}
	return usd
}

func hunyuanV3USD(generateType string) float64 {
	switch strings.ToLower(strings.TrimSpace(generateType)) {
	case "geometry":
		return hunyuanGeometryUSD
	case "lowpoly", "low_poly", "low poly":
		return hunyuanLowPolyUSD
	default:
		return hunyuanNormalUSD
	}
}

func hunyuanRapidT2DUSD(p ThreeDGenerationParams) float64 {
	usd := hunyuanRapidUSD
	if p.EnablePBRSet && p.EnablePBR {
		usd += hunyuanRapidPBRUSD
	}
	return usd
}

func hunyuanRapidI2DUSD(p ThreeDGenerationParams) float64 {
	usd := hunyuanRapidI2DBaseUSD
	if p.EnablePBRSet && p.EnablePBR {
		usd += hunyuanRapidPBRUSD
	}
	return usd
}

func rodinV25USD(tier, addons string) float64 {
	usd := rodinV25BaseUSD
	if strings.EqualFold(strings.TrimSpace(tier), "Gen-2.5-Extreme-High") {
		usd = rodinV25BaseUSD * 2
	}
	if strings.Contains(strings.ToLower(addons), "highpack") {
		usd += rodinHighPackUSD
	}
	return usd
}

func tripoH31TextureDeltas(modelID string, base int) map[string]int {
	ref := defaultThreeDUSD(modelID)
	return map[string]int{
		"true":  0,
		"false": usdToCoinsFromBase(modelID, base, ref, tripoH31USD(ThreeDGenerationParams{ModelID: modelID, Texture: false, TextureSet: true, GeometryQuality: "standard"})) - base,
	}
}

func tripoH31TextureQualityDeltas(modelID string, base int) map[string]int {
	ref := defaultThreeDUSD(modelID)
	return map[string]int{
		"standard": 0,
		"detailed": usdToCoinsFromBase(modelID, base, ref, tripoH31USD(ThreeDGenerationParams{ModelID: modelID, Texture: true, TextureSet: true, TextureQuality: "detailed", GeometryQuality: "standard"})) - base,
	}
}

func tripoH31GeometryQualityDeltas(modelID string, base int) map[string]int {
	ref := defaultThreeDUSD(modelID)
	return map[string]int{
		"standard": 0,
		"detailed": usdToCoinsFromBase(modelID, base, ref, tripoH31USD(ThreeDGenerationParams{ModelID: modelID, Texture: true, TextureSet: true, TextureQuality: "standard", GeometryQuality: "detailed"})) - base,
	}
}

func tripoH31QuadDeltas(modelID string, base int) map[string]int {
	ref := defaultThreeDUSD(modelID)
	return map[string]int{
		"false": 0,
		"true":  usdToCoinsFromBase(modelID, base, ref, tripoH31USD(ThreeDGenerationParams{ModelID: modelID, Texture: true, TextureSet: true, TextureQuality: "standard", GeometryQuality: "standard", Quad: true})) - base,
	}
}

func hunyuanGenerateTypeDeltas(modelID string, base int) map[string]int {
	ref := defaultThreeDUSD(modelID)
	return map[string]int{
		"Geometry": usdToCoinsFromBase(modelID, base, ref, hunyuanGeometryUSD) - base,
		"Normal":   0,
		"LowPoly":  usdToCoinsFromBase(modelID, base, ref, hunyuanLowPolyUSD) - base,
	}
}

func hunyuanRapidPBRDeltas(modelID string, base int) map[string]int {
	ref := defaultThreeDUSD(modelID)
	withPBR := ref + hunyuanRapidPBRUSD
	return map[string]int{
		"false": 0,
		"true":  usdToCoinsFromBase(modelID, base, ref, withPBR) - base,
	}
}

func rodinV25TierDeltas(modelID string, base int) map[string]int {
	ref := defaultThreeDUSD(modelID)
	deltas := map[string]int{}
	for _, tier := range []string{"Gen-2.5-Extreme-Low", "Gen-2.5-Low", "Gen-2.5-Medium", "Gen-2.5-High", "Gen-2.5-Extreme-High"} {
		deltas[tier] = usdToCoinsFromBase(modelID, base, ref, rodinV25USD(tier, "")) - base
	}
	deltas["Gen-2.5-Medium"] = 0
	return deltas
}

func rodinV25AddonsDeltas(modelID string, base int) map[string]int {
	ref := defaultThreeDUSD(modelID)
	return map[string]int{
		"":         0,
		"HighPack": usdToCoinsFromBase(modelID, base, ref, rodinV25USD("Gen-2.5-Medium", "HighPack")) - base,
	}
}
