package ai

import "strings"

const (
	seedanceVideoEditSlug      = "bytedance/seedance-2.0/video-edit"
	seedanceVideoEditTurboSlug = "bytedance/seedance-2.0/video-edit-turbo"
)

// selectSeedanceVideoEditSlug picks Edit or Edit Turbo from resolution and optional turbo flag.
func selectSeedanceVideoEditSlug(req VideoRequest) string {
	if req.TurboMode != nil {
		if *req.TurboMode {
			return seedanceVideoEditTurboSlug
		}
		return seedanceVideoEditSlug
	}

	switch strings.ToLower(strings.TrimSpace(req.Resolution)) {
	case "720p", "1080p":
		return seedanceVideoEditTurboSlug
	default:
		return seedanceVideoEditSlug
	}
}

func isUnifiedSeedanceVideoEdit(def mediaModelDef) bool {
	return def.ID == "seedance-v2-video-edit"
}

// wavespeedVideoEditModelID returns the model ID used for billing/response
// purposes after a video generation. Only seedance-v2-video-edit needs a
// suffix here — it has two WaveSpeed pricing tiers (standard vs turbo)
// sharing one catalog entry, so the turbo run must be tagged distinctly.
//
// PREVIOUSLY this appended "-edit" to the ID of every def.RequiresVideo
// model that wasn't one of the seedance-v2 special cases — e.g.
// "wan-2.7-edit" became "wan-2.7-edit-edit", "happyhorse-video-edit" became
// "happyhorse-video-edit-edit", etc. Those catalog IDs already end in
// "-edit"/"-extend" on their own (unlike seedance-v2-video-edit, which is a
// single catalog entry covering two distinct API endpoints), so the extra
// suffix just produced an ID that matched nothing in the price catalog —
// billing silently fell back to the generic per-category default price
// instead of whatever was actually configured for that model. Fixed by
// only special-casing the one model that genuinely needs it.
func wavespeedVideoEditModelID(def mediaModelDef, slug string) string {
	if isUnifiedSeedanceVideoEdit(def) && strings.Contains(slug, "turbo") {
		return def.ID + "-turbo"
	}
	return def.ID
}
