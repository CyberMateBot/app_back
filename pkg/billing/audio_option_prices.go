package billing

import (
	"strconv"
	"strings"
)

const (
	murekaPerSongUSD       = 0.045
	aceStepPerSecondUSD    = 0.0003
	qwenTTSMinUSD          = 0.005
	qwenTTSPer100USD       = 0.005
	qwenTTSClonePer100USD  = 0.05
	omnivoiceMinUSD        = 0.005
	omnivoicePer100USD     = 0.005
	elevenLabsPer1000USD   = 0.10
	elevenLabsMinChars     = 1000
	minimaxSpeechPer1000USD = 0.06
)

// AudioGenerationParams holds audio option values that affect generation cost.
type AudioGenerationParams struct {
	ModelID       string
	TextLength    int
	VoiceClone    bool
	Duration      int
	NumberOfSongs int
}

// AudioBillingModelID returns the catalog model ID used for pricing (e.g. Qwen3 voice clone).
func AudioBillingModelID(p AudioGenerationParams) string {
	modelID := strings.ToLower(strings.TrimSpace(p.ModelID))
	if modelID == "qwen3-tts" && p.VoiceClone {
		return "qwen3-tts-clone"
	}
	return modelID
}

// AudioOptionValuePrices returns coin surcharges per option value relative to the model base price.
func AudioOptionValuePrices(modelID, optionKey string) map[string]int {
	modelID = strings.ToLower(strings.TrimSpace(modelID))
	optionKey = strings.ToLower(strings.TrimSpace(optionKey))
	base := DefaultModelPrice(modelID, "audio")

	switch modelID {
	case "mureka-v9":
		if optionKey == "number_of_songs" {
			return murekaSongDeltas(modelID, base)
		}
	case "ace-step-1.5":
		if optionKey == "duration" {
			return aceStepDurationDeltas(modelID, base)
		}
	case "qwen3-tts", "qwen3-tts-clone", "omnivoice", "elevenlabs-v3", "minimax-speech-2.6":
		if optionKey == "text_length" {
			return ttsTextLengthDeltas(modelID, base)
		}
	}
	return nil
}

// AudioGenerationPrice returns total CyberCoins for an audio generation with the given options.
func AudioGenerationPrice(basePrice int, p AudioGenerationParams) int {
	effectiveID := AudioBillingModelID(p)
	if basePrice <= 0 {
		basePrice = DefaultModelPrice(effectiveID, "audio")
	}

	usd := audioGenerationUSD(effectiveID, p)
	if usd <= 0 {
		return basePrice
	}

	baseUSD := defaultAudioUSD(effectiveID)
	price := usdToCoinsFromBase(effectiveID, basePrice, baseUSD, usd)
	if price < 1 {
		return 1
	}
	return price
}

func audioGenerationUSD(modelID string, p AudioGenerationParams) float64 {
	chars := normalizeTextLength(p.TextLength)

	switch modelID {
	case "mureka-v9":
		songs := p.NumberOfSongs
		if songs < 1 {
			songs = 1
		}
		if songs > 3 {
			songs = 3
		}
		return murekaPerSongUSD * float64(songs)
	case "ace-step-1.5":
		duration := p.Duration
		if duration < 5 {
			duration = defaultAceStepDuration()
		}
		if duration > 240 {
			duration = 240
		}
		return aceStepPerSecondUSD * float64(duration)
	case "qwen3-tts":
		return qwenTTSUSD(chars)
	case "qwen3-tts-clone":
		return qwenTTSCloneUSD(chars)
	case "omnivoice":
		return omnivoiceUSD(chars)
	case "elevenlabs-v3":
		return elevenLabsUSD(chars)
	case "minimax-speech-2.6":
		return minimaxSpeechUSD(chars)
	default:
		return 0
	}
}

func defaultAudioUSD(modelID string) float64 {
	switch modelID {
	case "mureka-v9":
		return murekaPerSongUSD
	case "ace-step-1.5":
		return aceStepPerSecondUSD * float64(defaultAceStepDuration())
	case "qwen3-tts", "qwen3-tts-clone", "omnivoice":
		return qwenTTSMinUSD
	case "elevenlabs-v3":
		return elevenLabsPer1000USD
	case "minimax-speech-2.6":
		return minimaxSpeechUSD(defaultMinimaxReferenceChars())
	default:
		return 0
	}
}

func qwenTTSUSD(chars int) float64 {
	if chars < 100 {
		return qwenTTSMinUSD
	}
	return qwenTTSPer100USD * float64(chars) / 100.0
}

func qwenTTSCloneUSD(chars int) float64 {
	if chars < 100 {
		return qwenTTSMinUSD
	}
	return qwenTTSClonePer100USD * float64(chars) / 100.0
}

func omnivoiceUSD(chars int) float64 {
	if chars < 100 {
		return omnivoiceMinUSD
	}
	return omnivoicePer100USD * float64(chars) / 100.0
}

func elevenLabsUSD(chars int) float64 {
	billed := chars
	if billed < elevenLabsMinChars {
		billed = elevenLabsMinChars
	}
	return elevenLabsPer1000USD * float64(billed) / 1000.0
}

func minimaxSpeechUSD(chars int) float64 {
	return minimaxSpeechPer1000USD * float64(chars) / 1000.0
}

func normalizeTextLength(chars int) int {
	if chars < 1 {
		return 1
	}
	return chars
}

func defaultAceStepDuration() int {
	return 60
}

func defaultMinimaxReferenceChars() int {
	return 1000
}

func defaultTTSReferenceChars(modelID string) int {
	switch modelID {
	case "elevenlabs-v3", "minimax-speech-2.6":
		return 1000
	default:
		return 100
	}
}

func ttsTextLengthDeltas(modelID string, base int) map[string]int {
	refChars := defaultTTSReferenceChars(modelID)
	refUSD := audioGenerationUSD(modelID, AudioGenerationParams{ModelID: modelID, TextLength: refChars})
	tiers := []string{"50", "100", "500", "1000", "2000", "5000"}
	deltas := make(map[string]int, len(tiers))
	for _, tier := range tiers {
		chars, _ := strconv.Atoi(tier)
		usd := audioGenerationUSD(modelID, AudioGenerationParams{ModelID: modelID, TextLength: chars})
		deltas[tier] = usdToCoinsFromBase(modelID, base, refUSD, usd) - base
	}
	deltas[strconv.Itoa(refChars)] = 0
	return deltas
}

func murekaSongDeltas(modelID string, base int) map[string]int {
	refUSD := murekaPerSongUSD
	deltas := make(map[string]int, 3)
	for _, s := range []string{"1", "2", "3"} {
		n, _ := strconv.Atoi(s)
		usd := murekaPerSongUSD * float64(n)
		deltas[s] = usdToCoinsFromBase(modelID, base, refUSD, usd) - base
	}
	deltas["1"] = 0
	return deltas
}

func aceStepDurationDeltas(modelID string, base int) map[string]int {
	defaultDur := defaultAceStepDuration()
	refUSD := aceStepPerSecondUSD * float64(defaultDur)
	durations := []string{"30", "60", "120", "180", "240"}
	deltas := make(map[string]int, len(durations))
	for _, dStr := range durations {
		d, _ := strconv.Atoi(dStr)
		usd := aceStepPerSecondUSD * float64(d)
		deltas[dStr] = usdToCoinsFromBase(modelID, base, refUSD, usd) - base
	}
	deltas[strconv.Itoa(defaultDur)] = 0
	return deltas
}
