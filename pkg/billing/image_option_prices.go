package billing

import (
	"math"
	"strconv"
	"strings"
)

// ImageGenerationParams holds image option values that affect generation cost.
type ImageGenerationParams struct {
	ModelID     string
	Resolution  string
	Quality     string
	Size        string
	AspectRatio string
	NumImages   int
	WebSearch   bool
	ImageSearch bool
}

// WaveSpeed USD prices (https://wavespeed.ai/image-generator).

var nanoBananaWaveSpeedUSD = map[string]map[string]float64{
	"nano-banana-2": {
		"0.5k": 0.045,
		"1k":   0.07,
		"2k":   0.105,
		"4k":   0.14,
	},
	"nano-banana-pro": {
		"1k": 0.14,
		"2k": 0.14,
		"4k": 0.24,
	},
}

var gptImage2USD = map[string]map[string]float64{
	"low":    {"1k": 0.010, "2k": 0.020, "4k": 0.030},
	"medium": {"1k": 0.060, "2k": 0.120, "4k": 0.180},
	"high":   {"1k": 0.220, "2k": 0.440, "4k": 0.660},
}

// gpt-image-1.5 uses square (1k) vs landscape (2k) pricing tiers instead of literal resolution.
var gptImage15USD = map[string]map[string]float64{
	"low":    {"1k": 0.010, "2k": 0.015},
	"medium": {"1k": 0.040, "2k": 0.060},
	"high":   {"1k": 0.150, "2k": 0.250},
}

var grokImagineUSD = map[string]float64{
	"1k": 0.07,
	"2k": 0.09,
}

const (
	nanoBanana2WebSearchUSD   = 0.014
	nanoBanana2ImageSearchUSD = 0.014

	qwenImageBaseUSD    = 0.02
	qwenImageBasePixels = 1024 * 1024

	zImageBaseUSD    = 0.005
	zImageBasePixels = 1024 * 1024
)

// ImageOptionValuePrices returns coin surcharges per option value relative to the model base price.
// For interacting options (e.g. GPT Image quality × resolution) deltas are marginal at default values;
// the backend uses full lookup in ImageGenerationPrice.
func ImageOptionValuePrices(modelID, optionKey string) map[string]int {
	modelID = strings.ToLower(strings.TrimSpace(modelID))
	optionKey = strings.ToLower(strings.TrimSpace(optionKey))
	base := DefaultModelPrice(modelID, "image")

	switch modelID {
	case "nano-banana-2", "nano-banana-pro":
		switch optionKey {
		case "resolution":
			return resolutionPriceDeltas(modelID, base, nanoBananaWaveSpeedUSD[modelID], "1k")
		case "web_search":
			if modelID == "nano-banana-2" {
				return map[string]int{"true": usdToCoinsFromBase(modelID, base, nanoBananaWaveSpeedUSD[modelID]["1k"], nanoBanana2WebSearchUSD)}
			}
		case "image_search":
			if modelID == "nano-banana-2" {
				return map[string]int{"true": usdToCoinsFromBase(modelID, base, nanoBananaWaveSpeedUSD[modelID]["1k"], nanoBanana2ImageSearchUSD)}
			}
		}
	case "gpt-image-2":
		switch optionKey {
		case "resolution":
			return gptImage2ResolutionDeltas(base)
		case "quality":
			return gptImage2QualityDeltas(base)
		}
	case "gpt-image-1.5":
		switch optionKey {
		case "resolution":
			return gptImage15ResolutionDeltas(base)
		case "quality":
			return gptImage15QualityDeltas(base)
		}
	case "grok-imagine-edit":
		switch optionKey {
		case "resolution":
			return resolutionPriceDeltas(modelID, base, grokImagineUSD, "1k")
		case "num_images":
			return grokNumImagesDeltas(base)
		}
	case "qwen-image", "qwen-image-2512":
		if optionKey == "size" {
			return qwenImageSizeDeltas(modelID, base)
		}
	case "z-image-base", "z-image-turbo":
		if optionKey == "size" {
			return zImageSizeDeltas(modelID, base)
		}
	}
	return nil
}

// ImageGenerationPrice returns total CyberCoins for an image generation with the given options.
func ImageGenerationPrice(basePrice int, p ImageGenerationParams) int {
	modelID := strings.ToLower(strings.TrimSpace(p.ModelID))
	if basePrice <= 0 {
		basePrice = DefaultModelPrice(modelID, "image")
	}

	var usd float64
	switch modelID {
	case "nano-banana-2", "nano-banana-pro":
		table := nanoBananaWaveSpeedUSD[modelID]
		usd = table[normalizeNanoBananaResolution(modelID, p.Resolution)]
		if modelID == "nano-banana-2" {
			if p.WebSearch {
				usd += nanoBanana2WebSearchUSD
			}
			if p.ImageSearch {
				usd += nanoBanana2ImageSearchUSD
			}
		}
	case "gpt-image-2":
		usd = gptImage2USD[normalizeQuality(p.Quality)][normalizeResolution(p.Resolution)]
	case "gpt-image-1.5":
		tier := gptImage15SizeTier(p.Resolution, p.AspectRatio)
		usd = gptImage15USD[normalizeQuality(p.Quality)][tier]
	case "grok-imagine-edit":
		perImage := grokImagineUSD[normalizeGrokResolution(p.Resolution)]
		num := p.NumImages
		if num <= 0 {
			num = 1
		}
		if num > 4 {
			num = 4
		}
		usd = perImage * float64(num)
	case "qwen-image", "qwen-image-2512":
		usd = qwenImageUSD(p.Size)
	case "z-image-base", "z-image-turbo":
		usd = zImageUSD(p.Size)
	default:
		return basePrice
	}

	baseUSD := defaultUSDForModel(modelID, p)
	price := usdToCoinsFromBase(modelID, basePrice, baseUSD, usd)
	if price < 1 {
		return 1
	}
	return price
}

func defaultUSDForModel(modelID string, p ImageGenerationParams) float64 {
	switch modelID {
	case "nano-banana-2", "nano-banana-pro":
		return nanoBananaWaveSpeedUSD[modelID]["1k"]
	case "gpt-image-2":
		return gptImage2USD["medium"]["1k"]
	case "gpt-image-1.5":
		return gptImage15USD["medium"]["1k"]
	case "grok-imagine-edit":
		return grokImagineUSD["1k"]
	case "qwen-image", "qwen-image-2512":
		return qwenImageUSD("1024*1024")
	case "z-image-base", "z-image-turbo":
		return zImageUSD("1024*1024")
	default:
		return 0
	}
}

func gptImage2ResolutionDeltas(base int) map[string]int {
	ref := gptImage2USD["medium"]["1k"]
	deltas := make(map[string]int, 3)
	for res, usd := range gptImage2USD["medium"] {
		deltas[res] = usdToCoinsFromBase("gpt-image-2", base, ref, usd) - base
	}
	deltas["1k"] = 0
	return deltas
}

func gptImage2QualityDeltas(base int) map[string]int {
	ref := gptImage2USD["medium"]["1k"]
	deltas := make(map[string]int, 3)
	for q, table := range gptImage2USD {
		deltas[q] = usdToCoinsFromBase("gpt-image-2", base, ref, table["1k"]) - base
	}
	deltas["medium"] = 0
	return deltas
}

func gptImage15ResolutionDeltas(base int) map[string]int {
	ref := gptImage15USD["medium"]["1k"]
	deltas := map[string]int{
		"1k": usdToCoinsFromBase("gpt-image-1.5", base, ref, gptImage15USD["medium"]["1k"]) - base,
		"2k": usdToCoinsFromBase("gpt-image-1.5", base, ref, gptImage15USD["medium"]["2k"]) - base,
		"4k": usdToCoinsFromBase("gpt-image-1.5", base, ref, gptImage15USD["medium"]["2k"]) - base,
	}
	deltas["1k"] = 0
	return deltas
}

func gptImage15QualityDeltas(base int) map[string]int {
	ref := gptImage15USD["medium"]["1k"]
	deltas := make(map[string]int, 3)
	for q, table := range gptImage15USD {
		deltas[q] = usdToCoinsFromBase("gpt-image-1.5", base, ref, table["1k"]) - base
	}
	deltas["medium"] = 0
	return deltas
}

func grokNumImagesDeltas(base int) map[string]int {
	ref := grokImagineUSD["1k"]
	perImage := usdToCoinsFromBase("grok-imagine-edit", base, ref, ref)
	return map[string]int{
		"2": perImage - base,
		"3": perImage*2 - base,
		"4": perImage*3 - base,
	}
}

func qwenImageSizeDeltas(modelID string, base int) map[string]int {
	ref := qwenImageUSD("1024*1024")
	deltas := map[string]int{
		"1024*1024":  0,
		"1024x1024":  0,
		"1328*1328":  usdToCoinsFromBase(modelID, base, ref, qwenImageUSD("1328*1328")) - base,
	}
	return deltas
}

func zImageSizeDeltas(modelID string, base int) map[string]int {
	ref := zImageUSD("1024*1024")
	// Only default size in options; deltas stay 0 unless more sizes are added.
	_ = ref
	return map[string]int{
		"1024*1024": 0,
		"1024x1024": 0,
	}
}

func resolutionPriceDeltas(modelID string, base int, table map[string]float64, defaultRes string) map[string]int {
	baseUSD := table[defaultRes]
	deltas := make(map[string]int, len(table))
	for res, usd := range table {
		deltas[res] = usdToCoinsFromBase(modelID, base, baseUSD, usd) - base
	}
	deltas[defaultRes] = 0
	return deltas
}

func qwenImageUSD(size string) float64 {
	pixels, ok := sizePixelArea(size)
	if !ok || pixels <= 0 {
		return qwenImageBaseUSD
	}
	return qwenImageBaseUSD * float64(pixels) / float64(qwenImageBasePixels)
}

func zImageUSD(size string) float64 {
	pixels, ok := sizePixelArea(size)
	if !ok || pixels <= 0 {
		return zImageBaseUSD
	}
	return zImageBaseUSD * float64(pixels) / float64(zImageBasePixels)
}

func usdToCoinsFromBase(modelID string, baseCoins int, baseUSD, usd float64) int {
	if baseUSD <= 0 || usd <= 0 {
		return int(math.Round(usd * 3.5 * 100))
	}
	scale := float64(baseCoins) / baseUSD
	return int(math.Round(usd * scale))
}

func normalizeQuality(quality string) string {
	switch strings.ToLower(strings.TrimSpace(quality)) {
	case "low", "high":
		return strings.ToLower(strings.TrimSpace(quality))
	default:
		return "medium"
	}
}

func normalizeResolution(resolution string) string {
	switch strings.ToLower(strings.TrimSpace(resolution)) {
	case "2k", "4k":
		return strings.ToLower(strings.TrimSpace(resolution))
	default:
		return "1k"
	}
}

func normalizeGrokResolution(resolution string) string {
	if strings.ToLower(strings.TrimSpace(resolution)) == "2k" {
		return "2k"
	}
	return "1k"
}

func normalizeNanoBananaResolution(modelID, resolution string) string {
	resolution = strings.ToLower(strings.TrimSpace(resolution))
	switch modelID {
	case "nano-banana-2":
		switch resolution {
		case "512px", "512", "0.5k":
			return "0.5k"
		case "2k", "4k":
			return resolution
		default:
			return "1k"
		}
	case "nano-banana-pro":
		switch resolution {
		case "2k", "4k":
			return resolution
		default:
			return "1k"
		}
	default:
		return resolution
	}
}

func gptImage15SizeTier(resolution, aspectRatio string) string {
	res := normalizeResolution(resolution)
	if res == "2k" || res == "4k" {
		return "2k"
	}
	ar := strings.TrimSpace(aspectRatio)
	if ar == "" || ar == "1:1" {
		return "1k"
	}
	return "2k"
}

func sizePixelArea(size string) (int, bool) {
	size = strings.TrimSpace(size)
	size = strings.ReplaceAll(size, "x", "*")
	parts := strings.Split(size, "*")
	if len(parts) != 2 {
		return 0, false
	}
	w, errW := strconv.Atoi(strings.TrimSpace(parts[0]))
	h, errH := strconv.Atoi(strings.TrimSpace(parts[1]))
	if errW != nil || errH != nil || w <= 0 || h <= 0 {
		return 0, false
	}
	return w * h, true
}
