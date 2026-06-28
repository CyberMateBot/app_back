package billing

import (
	"strconv"
	"strings"
)

// VideoGenerationParams holds video option values that affect generation cost.
type VideoGenerationParams struct {
	ModelID       string
	Duration      int
	ExtendBy      int
	Resolution    string
	Sound         bool
	GenerateAudio bool
	TurboMode     bool
}

var klingPerSecondUSD = map[string]float64{
	"kling-v3-std": 0.084,
	"kling-v3-pro": 0.112,
	"kling-v3-4k":  0.168,
}

var wanPerSecondUSD = map[string]float64{
	"480p": 0.05,
	"480P": 0.05,
	"720p": 0.10,
	"720P": 0.10,
	"1080p": 0.15,
	"1080P": 0.15,
}

var viduPerSecondUSD = map[string]float64{
	"540p":  0.07,
	"720p":  0.15,
	"1080p": 0.16,
}

var seedanceV2EditPerSecondUSD = map[string]map[string]float64{
	"standard": {
		"480p":  0.075,
		"720p":  0.15,
		"1080p": 0.375,
	},
	"turbo": {
		"720p":  0.085,
		"1080p": 0.095,
	},
}

// VideoOptionValuePrices returns coin surcharges per option value relative to the model base price.
func VideoOptionValuePrices(modelID, optionKey string) map[string]int {
	modelID = strings.ToLower(strings.TrimSpace(modelID))
	optionKey = strings.ToLower(strings.TrimSpace(optionKey))
	base := DefaultModelPrice(modelID, "video")

	switch optionKey {
	case "duration":
		return videoDurationDeltas(modelID, base)
	case "resolution":
		return videoResolutionDeltas(modelID, base)
	case "sound":
		if isKlingModel(modelID) {
			return klingSoundDeltas(modelID, base)
		}
	case "generate_audio":
		if isSeedance15Model(modelID) {
			return seedance15AudioDeltas(modelID, base)
		}
	case "turbo_mode":
		if modelID == "seedance-v2-video-edit" {
			return seedanceV2TurboDeltas(modelID, base)
		}
	case "extend_by":
		if modelID == "happyhorse-video-extend" || modelID == "veo-3.1-extend" {
			return videoExtendByDeltas(modelID, base)
		}
	}
	return nil
}

// VideoBillingModelID returns the catalog model ID used for pricing (e.g. Kling tier by resolution).
func VideoBillingModelID(p VideoGenerationParams) string {
	return videoBillingModelID(p)
}

// VideoGenerationPrice returns total CyberCoins for a video generation with the given options.
func VideoGenerationPrice(basePrice int, p VideoGenerationParams) int {
	modelID := strings.ToLower(strings.TrimSpace(p.ModelID))
	effectiveID := videoBillingModelID(p)
	if basePrice <= 0 || effectiveID != modelID {
		basePrice = DefaultModelPrice(effectiveID, "video")
	}

	usd := videoGenerationUSD(p)
	if usd <= 0 {
		return basePrice
	}

	baseUSD := defaultVideoUSD(effectiveID, p)
	price := usdToCoinsFromBase(effectiveID, basePrice, baseUSD, usd)
	if price < 1 {
		return 1
	}
	return price
}

func videoBillingModelID(p VideoGenerationParams) string {
	modelID := strings.ToLower(strings.TrimSpace(p.ModelID))
	if isKlingModel(modelID) {
		return klingEffectiveModel(modelID, p.Resolution)
	}
	return modelID
}

func videoGenerationUSD(p VideoGenerationParams) float64 {
	modelID := videoBillingModelID(p)
	duration := normalizeVideoDuration(p.Duration, defaultVideoDuration(modelID))

	switch {
	case isKlingModel(modelID):
		usd := klingPerSecondUSD[modelID] * float64(duration)
		if p.Sound {
			usd *= 1.5
		}
		return usd
	case isSeedance15Model(modelID):
		return seedance15PerSecond(p.Resolution, p.GenerateAudio) * float64(duration)
	case modelID == "seedance-v1-pro-i2v":
		return 0.06 * float64(duration)
	case modelID == "seedance-v2-video-edit":
		tier := "standard"
		if p.TurboMode || isSeedanceTurboResolution(p.Resolution) {
			tier = "turbo"
		}
		res := normalizeSeedanceResolution(p.Resolution, tier)
		rate := seedanceV2EditPerSecondUSD[tier][res]
		// Billed for input+output; estimate 2× output duration when input length is unknown.
		return rate * float64(duration) * 2
	case modelID == "seedance-v2-video-extend":
		return 0.15 * float64(duration) * 2
	case isWANModel(modelID):
		res := normalizeWANResolution(p.Resolution, modelID)
		return wanRateUSD(modelID, res) * float64(duration)
	case isHappyHorseModel(modelID):
		if modelID == "happyhorse-video-extend" {
			extend := p.ExtendBy
			if extend <= 0 {
				extend = duration
			}
			return happyHorseUSD("720p", extend)
		}
		return happyHorseUSD(p.Resolution, duration)
	case modelID == "veo-3.1-extend":
		return 1.05
	case modelID == "vidu-q3-i2v-spicy":
		res := normalizeViduResolution(p.Resolution)
		return viduPerSecondUSD[res] * float64(duration)
	case strings.HasPrefix(modelID, "hailuo-2.3"):
		return hailuoUSD(modelID, duration)
	default:
		return 0
	}
}

func defaultVideoUSD(modelID string, p VideoGenerationParams) float64 {
	duration := defaultVideoDuration(modelID)
	switch {
	case isKlingModel(modelID):
		return klingPerSecondUSD[modelID] * float64(duration)
	case isSeedance15Model(modelID):
		return seedance15PerSecond(defaultVideoResolution(modelID), defaultSeedance15Audio(modelID)) * float64(duration)
	case modelID == "seedance-v1-pro-i2v":
		return 0.06 * float64(duration)
	case modelID == "seedance-v2-video-edit":
		p.Resolution = defaultVideoResolution(modelID)
		p.Duration = duration
		return videoGenerationUSD(VideoGenerationParams{ModelID: modelID, Duration: duration, Resolution: p.Resolution})
	case modelID == "seedance-v2-video-extend":
		return videoGenerationUSD(VideoGenerationParams{ModelID: modelID, Duration: duration, Resolution: defaultVideoResolution(modelID)})
	case isWANModel(modelID):
		res := normalizeWANResolution(defaultVideoResolution(modelID), modelID)
		return wanRateUSD(modelID, res) * float64(duration)
	case isHappyHorseModel(modelID):
		if modelID == "happyhorse-video-extend" {
			return happyHorseUSD("720p", 5)
		}
		return happyHorseUSD(defaultVideoResolution(modelID), duration)
	case modelID == "veo-3.1-extend":
		return 1.05
	case modelID == "vidu-q3-i2v-spicy":
		return viduPerSecondUSD["720p"] * float64(duration)
	case strings.HasPrefix(modelID, "hailuo-2.3"):
		return hailuoUSD(modelID, duration)
	default:
		return 0
	}
}

func videoDurationDeltas(modelID string, base int) map[string]int {
	durations := videoDurationOptions(modelID)
	if len(durations) == 0 {
		return nil
	}
	defaultDur := defaultVideoDuration(modelID)
	refUSD := videoGenerationUSD(VideoGenerationParams{ModelID: modelID, Duration: defaultDur, Resolution: defaultVideoResolution(modelID)})
	deltas := make(map[string]int, len(durations))
	for _, dStr := range durations {
		d, err := strconv.Atoi(dStr)
		if err != nil {
			continue
		}
		usd := videoGenerationUSD(VideoGenerationParams{
			ModelID:       modelID,
			Duration:      d,
			Resolution:    defaultVideoResolution(modelID),
			GenerateAudio: defaultSeedance15Audio(modelID),
		})
		deltas[dStr] = usdToCoinsFromBase(modelID, base, refUSD, usd) - base
	}
	deltas[strconv.Itoa(defaultDur)] = 0
	return deltas
}

func videoResolutionDeltas(modelID string, base int) map[string]int {
	resolutions := videoResolutionOptions(modelID)
	if len(resolutions) == 0 {
		return nil
	}
	defaultRes := defaultVideoResolution(modelID)
	defaultDur := defaultVideoDuration(modelID)
	refUSD := videoGenerationUSD(VideoGenerationParams{
		ModelID: modelID, Duration: defaultDur, Resolution: defaultRes, GenerateAudio: defaultSeedance15Audio(modelID),
	})
	deltas := make(map[string]int, len(resolutions))
	for _, res := range resolutions {
		usd := videoGenerationUSD(VideoGenerationParams{
			ModelID:       modelID,
			Duration:      defaultDur,
			Resolution:    res,
			GenerateAudio: defaultSeedance15Audio(modelID),
		})
		deltas[res] = usdToCoinsFromBase(modelID, base, refUSD, usd) - base
	}
	if defaultRes != "" {
		deltas[defaultRes] = 0
	}
	return deltas
}

func klingSoundDeltas(modelID string, base int) map[string]int {
	effective := klingEffectiveModel(modelID, defaultVideoResolution(modelID))
	dur := defaultVideoDuration(effective)
	ref := klingPerSecondUSD[effective] * float64(dur)
	withSound := ref * 1.5
	return map[string]int{
		"true": usdToCoinsFromBase(effective, base, ref, withSound) - base,
	}
}

func seedance15AudioDeltas(modelID string, base int) map[string]int {
	dur := defaultVideoDuration(modelID)
	res := defaultVideoResolution(modelID)
	ref := seedance15PerSecond(res, true) * float64(dur)
	without := seedance15PerSecond(res, false) * float64(dur)
	return map[string]int{
		"true":  0,
		"false": usdToCoinsFromBase(modelID, base, ref, without) - base,
	}
}

func seedanceV2TurboDeltas(modelID string, base int) map[string]int {
	dur := defaultVideoDuration(modelID)
	ref := videoGenerationUSD(VideoGenerationParams{ModelID: modelID, Duration: dur, Resolution: "720p", TurboMode: true})
	standard := videoGenerationUSD(VideoGenerationParams{ModelID: modelID, Duration: dur, Resolution: "480p", TurboMode: false})
	return map[string]int{
		"true":  usdToCoinsFromBase(modelID, base, ref, ref) - base,
		"false": usdToCoinsFromBase(modelID, base, ref, standard) - base,
	}
}

func videoExtendByDeltas(modelID string, base int) map[string]int {
	if modelID == "veo-3.1-extend" {
		return nil
	}
	ref := videoGenerationUSD(VideoGenerationParams{ModelID: modelID, Duration: 5, ExtendBy: 5})
	deltas := map[string]int{}
	for _, d := range []string{"3", "5", "10"} {
		val, _ := strconv.Atoi(d)
		usd := videoGenerationUSD(VideoGenerationParams{ModelID: modelID, Duration: 5, ExtendBy: val})
		deltas[d] = usdToCoinsFromBase(modelID, base, ref, usd) - base
	}
	deltas["5"] = 0
	return deltas
}

func seedance15PerSecond(resolution string, generateAudio bool) float64 {
	if strings.EqualFold(resolution, "1080p") {
		if generateAudio {
			return 0.06
		}
		return 0.03
	}
	if generateAudio {
		return 0.04
	}
	return 0.02
}

func happyHorseUSD(resolution string, duration int) float64 {
	if duration <= 0 {
		duration = 5
	}
	mult := 1.0
	if strings.EqualFold(resolution, "1080p") {
		mult = 2
	}
	return 0.70 * mult * float64(duration) / 5.0
}

func hailuoUSD(modelID string, duration int) float64 {
	switch modelID {
	case "hailuo-2.3-t2v":
		if duration >= 10 {
			return 0.56
		}
		return 0.23
	case "hailuo-2.3-i2v-fast":
		return 0.19
	case "hailuo-2.3-i2v-pro":
		return 0.49
	default:
		return 0.23
	}
}

func wanRateUSD(modelID, resolution string) float64 {
	if strings.HasPrefix(modelID, "wan-2.7") {
		switch resolution {
		case "1080P", "1080p":
			return 0.15
		default:
			return 0.10
		}
	}
	return wanPerSecondUSD[resolution]
}

func klingEffectiveModel(modelID, resolution string) string {
	if !isKlingModel(modelID) {
		return modelID
	}
	switch strings.ToLower(strings.TrimSpace(resolution)) {
	case "4k":
		return "kling-v3-4k"
	case "1080p":
		return "kling-v3-pro"
	case "720p":
		return "kling-v3-std"
	default:
		return modelID
	}
}

func isKlingModel(modelID string) bool {
	return strings.HasPrefix(modelID, "kling-v3")
}

func isSeedance15Model(modelID string) bool {
	return strings.HasPrefix(modelID, "seedance-v1.5")
}

func isWANModel(modelID string) bool {
	return strings.HasPrefix(modelID, "wan-")
}

func isHappyHorseModel(modelID string) bool {
	return strings.HasPrefix(modelID, "happyhorse-")
}

func isSeedanceTurboResolution(resolution string) bool {
	switch strings.ToLower(strings.TrimSpace(resolution)) {
	case "720p", "1080p":
		return true
	default:
		return false
	}
}

func normalizeSeedanceResolution(resolution, tier string) string {
	res := strings.ToLower(strings.TrimSpace(resolution))
	if res == "" {
		if tier == "turbo" {
			return "720p"
		}
		return "480p"
	}
	if tier == "turbo" && res == "480p" {
		return "720p"
	}
	return res
}

func normalizeWANResolution(resolution, modelID string) string {
	res := strings.TrimSpace(resolution)
	if res == "" {
		res = defaultVideoResolution(modelID)
	}
	switch strings.ToUpper(res) {
	case "480P", "480p":
		return "480P"
	case "1080P", "1080p":
		return "1080P"
	default:
		return "720P"
	}
}

func normalizeViduResolution(resolution string) string {
	res := strings.ToLower(strings.TrimSpace(resolution))
	if res == "" {
		return "720p"
	}
	return res
}

func normalizeVideoDuration(duration, fallback int) int {
	if duration > 0 {
		return duration
	}
	if fallback > 0 {
		return fallback
	}
	return 5
}

func defaultVideoDuration(modelID string) int {
	switch {
	case strings.HasPrefix(modelID, "hailuo-2.3-t2v"):
		return 6
	case strings.HasPrefix(modelID, "vidu-"):
		return 5
	case strings.HasPrefix(modelID, "wan-2.7-flf"), strings.HasPrefix(modelID, "wan-2.7-grid"):
		return 5
	default:
		return 5
	}
}

func defaultVideoResolution(modelID string) string {
	switch {
	case strings.HasPrefix(modelID, "kling-v3-4k"):
		return "4k"
	case strings.HasPrefix(modelID, "kling-v3-pro"):
		return "1080p"
	case strings.HasPrefix(modelID, "kling-v3"):
		return "720p"
	case strings.HasPrefix(modelID, "seedance-v1.5"), strings.HasPrefix(modelID, "happyhorse-"), strings.HasPrefix(modelID, "vidu-"):
		return "720p"
	case strings.HasPrefix(modelID, "seedance-v2"):
		return "720p"
	case strings.HasPrefix(modelID, "wan-2.7"):
		return "720P"
	case strings.HasPrefix(modelID, "wan-"):
		return "720P"
	case modelID == "veo-3.1-extend":
		return "1080p"
	default:
		return ""
	}
}

func defaultSeedance15Audio(modelID string) bool {
	return isSeedance15Model(modelID)
}

func videoDurationOptions(modelID string) []string {
	switch {
	case isKlingModel(modelID):
		return []string{"3", "5", "10", "15"}
	case modelID == "seedance-v1-pro-i2v":
		return []string{"2", "5", "8", "10", "12"}
	case isSeedance15Model(modelID):
		return []string{"4", "5", "8", "10", "12"}
	case modelID == "seedance-v2-video-edit", modelID == "seedance-v2-video-extend":
		return []string{"4", "5", "8", "10", "12", "15"}
	case modelID == "wan-2.5-t2v", modelID == "wan-2.7-t2v":
		return []string{"2", "5", "10", "15"}
	case modelID == "wan-2.6-i2v", modelID == "wan-2.2-spicy-i2v":
		return []string{"5", "8"}
	case modelID == "wan-2.7-flf", modelID == "wan-2.7-grid":
		return []string{"5", "10"}
	case isHappyHorseModel(modelID) && modelID != "happyhorse-video-extend":
		return []string{"3", "5", "10", "15"}
	case modelID == "vidu-q3-i2v-spicy":
		return []string{"1", "5", "10", "16"}
	default:
		return nil
	}
}

func videoResolutionOptions(modelID string) []string {
	switch {
	case isKlingModel(modelID):
		return []string{"720p", "1080p", "4k"}
	case isSeedance15Model(modelID), isHappyHorseModel(modelID), modelID == "seedance-v2-video-extend":
		return []string{"720p", "1080p"}
	case modelID == "seedance-v2-video-edit":
		return []string{"480p", "720p", "1080p"}
	case modelID == "wan-2.5-t2v", modelID == "wan-2.7-t2v":
		return []string{"480P", "720P", "1080P"}
	case modelID == "wan-2.6-i2v", modelID == "wan-2.2-spicy-i2v":
		return []string{"480P", "720P", "1080p"}
	case modelID == "vidu-q3-i2v-spicy":
		return []string{"540p", "720p", "1080p"}
	case modelID == "veo-3.1-extend":
		return []string{"720p", "1080p"}
	default:
		return nil
	}
}
