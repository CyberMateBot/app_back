package billing

import "strings"

// Subscription plan identifiers (lowest → highest tier).
const (
	PlanFree  = "free"
	PlanBasic = "basic"
	PlanPro   = "pro"
	PlanMax   = "max"
	PlanUltra = "ultra"
)

// planRanks maps a plan id to an ordered rank. Higher rank unlocks more models.
var planRanks = map[string]int{
	PlanFree:  0,
	PlanBasic: 1,
	PlanPro:   2,
	PlanMax:   3,
	PlanUltra: 4,
}

// PlanOrder lists plan ids from lowest to highest tier.
var PlanOrder = []string{PlanFree, PlanBasic, PlanPro, PlanMax, PlanUltra}

// PlanRank returns the numeric rank of a plan id (free=0 by default).
func PlanRank(planID string) int {
	if r, ok := planRanks[normalizePlanID(planID)]; ok {
		return r
	}
	return 0
}

// PlanIDForRank returns the canonical plan id for a rank (clamped to range).
func PlanIDForRank(rank int) string {
	if rank < 0 {
		rank = 0
	}
	if rank >= len(PlanOrder) {
		rank = len(PlanOrder) - 1
	}
	return PlanOrder[rank]
}

// IsValidPlanID reports whether the given id is a known subscription plan.
func IsValidPlanID(planID string) bool {
	_, ok := planRanks[normalizePlanID(planID)]
	return ok
}

func normalizePlanID(planID string) string {
	return strings.ToLower(strings.TrimSpace(planID))
}

// modelMinPlanRank is the canonical access table: the minimum plan rank required
// to use each model. Models not listed fall back to categoryMinPlanRank.
//
// The principle: free (rank 0) users get only the cheapest models, and only a few
// per category (no video, no 3D). Each higher tier unlocks progressively more and
// the most expensive premium models.
var modelMinPlanRank = map[string]int{
	// ---- Text ----
	// free (cheapest only)
	"yandexgpt": 0, "gpt-oss-20b": 0, "deepseek-chat": 0,
	"gpt-4o-mini": 1,
	// basic
	"gpt-oss-120b": 1, "qwen3-235b": 1, "deepseek-v4-flash": 1, "deepseek-v3.2": 1,
	"deepseek-chat-v3-0324": 1, "gpt-4.1-nano": 1, "gpt-4.1-mini": 1, "gpt-5.4-mini": 1,
	"claude-haiku-4.5": 1, "gemini-2.5-flash": 1, "o4-mini": 1,
	// pro
	"qwen3.6-35b": 2, "deepseek-v4": 2, "deepseek-r1": 2, "deepseek-v3.2-exp": 2,
	"gpt-4.1": 2, "gpt-5.4": 2, "claude-sonnet-4.5": 2, "o3-mini": 2,
	// max
	"gpt-4o": 3, "gemini-2.5-pro": 3, "claude-opus-4.7": 3, "o3": 3,
	// ultra
	"claude-opus-4.8": 4, "o1": 4, "gpt-5.5": 4,

	// ---- Image ----
	"flux-dev": 0,
	"alice-ai-art": 1,
	"nano-banana": 1,
	"gpt-image-1.5": 2, "gpt-image-2": 2, "nano-banana-2": 2,
	"nano-banana-pro": 3,

	// ---- Video (no free tier) ----
	"kling-v3-std": 1,
	"kling-v3-pro": 2, "seedance-v1-pro-i2v": 2, "seedance-v1.5-i2v-fast": 2,
	"seedance-v1.5-t2v-fast": 2, "seedance-v1.5-i2v-spicy": 2,
	"kling-v3-4k": 3, "seedance-v2-video-edit": 3, "seedance-v2-video-extend": 3,

	// ---- Audio ----
	"omnivoice": 0, "minimax-speech-2.6": 0, "qwen3-tts": 0,
	"elevenlabs-v3": 1,
	"mureka-v9": 2, "mureka": 2, "ace-step-1.5": 2,

	// ---- 3D (no free tier) ----
	"hunyuan3d-v3.1-rapid": 1, "hunyuan3d-v3-t2d": 1,
	"tripo3d-v2.5-i2d": 2, "tripo3d-v2.5-multiview": 2, "meshy6-t2d": 2,
	"tripo3d-h3.1-t2d": 3, "tripo3d-h3.1-i2d": 3, "rodin-v2-i2d": 3, "rodin-v2.5-i2d": 3,
}

// categoryMinPlanRank is the fallback minimum rank when a model id is unknown.
// Video and 3D have no free tier.
var categoryMinPlanRank = map[string]int{
	"text":  0,
	"image": 1,
	"video": 1,
	"audio": 0,
	"3d":    1,
}

// MinPlanRankForModel returns the minimum subscription rank required to use the model.
func MinPlanRankForModel(modelID, category string) int {
	id := strings.ToLower(strings.TrimSpace(modelID))
	if id != "" {
		if rank, ok := modelMinPlanRank[id]; ok {
			return rank
		}
	}
	if rank, ok := categoryMinPlanRank[strings.ToLower(strings.TrimSpace(category))]; ok {
		return rank
	}
	return 1
}

// MinPlanForModel returns the minimum plan id required to use the model.
func MinPlanForModel(modelID, category string) string {
	return PlanIDForRank(MinPlanRankForModel(modelID, category))
}

// PlanUnlocksModel reports whether a user on planID may use the model.
func PlanUnlocksModel(planID, modelID, category string) bool {
	return PlanRank(planID) >= MinPlanRankForModel(modelID, category)
}
