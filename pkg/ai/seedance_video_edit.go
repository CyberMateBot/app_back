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

func wavespeedVideoEditModelID(def mediaModelDef, slug string) string {
	if isUnifiedSeedanceVideoEdit(def) && strings.Contains(slug, "turbo") {
		return def.ID + "-turbo"
	}
	if def.RequiresVideo && def.ID == "seedance-v2-video-extend" {
		return def.ID
	}
	if def.RequiresVideo && isUnifiedSeedanceVideoEdit(def) {
		return def.ID
	}
	if def.RequiresVideo {
		return def.ID + "-edit"
	}
	return def.ID
}
