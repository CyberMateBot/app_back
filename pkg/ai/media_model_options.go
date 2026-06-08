package ai

import "strings"

// MediaOption describes a configurable generation parameter exposed to clients.
type MediaOption struct {
	Key     string   `json:"key"`
	Label   string   `json:"label,omitempty"`
	Type    string   `json:"type"`
	Values  []string `json:"values"`
	Default string   `json:"default,omitempty"`
}

func imageModelOptions(id string) []MediaOption {
	switch id {
	case "nano-banana":
		return []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"1:1", "16:9", "9:16", "4:3", "3:2"}, Default: "1:1"},
			{Key: "output_format", Type: "select", Values: []string{"png", "jpeg", "webp"}, Default: "png"},
		}
	case "nano-banana-pro":
		return []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"1:1", "3:2", "4:3", "16:9", "21:9"}, Default: "16:9"},
			{Key: "resolution", Type: "select", Values: []string{"1k", "2k", "4k"}, Default: "1k"},
			{Key: "output_format", Type: "select", Values: []string{"png", "jpeg", "webp"}, Default: "png"},
		}
	case "nano-banana-2":
		return []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"1:1", "2:3", "3:2", "3:4", "4:3", "4:5", "5:4", "9:16", "16:9", "21:9"}, Default: "16:9"},
			{Key: "resolution", Type: "select", Values: []string{"0.5k", "1k", "2k", "4k"}, Default: "1k"},
			{Key: "output_format", Type: "select", Values: []string{"png", "jpeg"}, Default: "png"},
		}
	case "gpt-image-2", "gpt-image-1.5":
		defAR := "16:9"
		if id == "gpt-image-1.5" {
			defAR = "1:1"
		}
		return []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"1:1", "3:2", "2:3", "3:4", "4:3", "4:5", "5:4", "9:16", "16:9", "21:9"}, Default: defAR},
			{Key: "resolution", Type: "select", Values: []string{"1k", "2k", "4k"}, Default: "1k"},
			{Key: "quality", Type: "select", Values: []string{"low", "medium", "high"}, Default: "medium"},
			{Key: "output_format", Type: "select", Values: []string{"png"}, Default: "png"},
		}
	case "flux-dev":
		return []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"1:1", "16:9", "9:16"}, Default: "1:1"},
		}
	default:
		return nil
	}
}

func videoModelOptions(id string) []MediaOption {
	switch id {
	case "kling-v3-std", "kling-v3-pro":
		opts := []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"16:9", "9:16", "1:1"}, Default: "16:9"},
			{Key: "duration", Type: "select", Values: klingDurationOptionValues(), Default: "5"},
			{Key: "quality_tier", Type: "select", Values: []string{"std", "pro"}, Default: qualityTierDefault(id)},
			{Key: "negative_prompt", Type: "text", Values: nil, Default: ""},
			{Key: "camera_movement", Type: "select", Values: []string{
				"auto", "simple", "down_back", "forward_up", "right_turn_forward", "left_turn_forward",
			}, Default: "auto"},
			{Key: "camera_horizontal", Type: "range", Values: []string{"-10", "10"}, Default: "0"},
			{Key: "camera_vertical", Type: "range", Values: []string{"-10", "10"}, Default: "0"},
			{Key: "camera_pan", Type: "range", Values: []string{"-10", "10"}, Default: "0"},
			{Key: "camera_tilt", Type: "range", Values: []string{"-10", "10"}, Default: "0"},
			{Key: "camera_roll", Type: "range", Values: []string{"-10", "10"}, Default: "0"},
			{Key: "camera_zoom", Type: "range", Values: []string{"-10", "10"}, Default: "0"},
		}
		if klingModelSupportsSound(id) {
			opts = append(opts, MediaOption{
				Key: "sound", Type: "boolean", Values: []string{"false", "true"}, Default: "false",
			})
		}
		return opts
	case "seedance-v1-pro-i2v":
		return []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"21:9", "16:9", "4:3", "1:1", "3:4", "9:16"}, Default: "16:9"},
			{Key: "duration", Type: "select", Values: []string{"2", "5", "8", "10", "12"}, Default: "5"},
			{Key: "camera_fixed", Type: "boolean", Values: []string{"false", "true"}, Default: "false"},
		}
	case "seedance-v1.5-i2v-fast", "seedance-v1.5-t2v-fast", "seedance-v1.5-i2v-spicy":
		return []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"16:9", "9:16", "4:3", "1:1", "3:4", "21:9"}, Default: "16:9"},
			{Key: "duration", Type: "select", Values: []string{"4", "5", "8", "10", "12"}, Default: "5"},
			{Key: "resolution", Type: "select", Values: []string{"720p", "1080p"}, Default: "720p"},
			{Key: "generate_audio", Type: "boolean", Values: []string{"true", "false"}, Default: "true"},
			{Key: "camera_fixed", Type: "boolean", Values: []string{"false", "true"}, Default: "false"},
		}
	case "seedance-v2-video-edit":
		return []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"16:9", "9:16", "4:3", "3:4", "1:1", "21:9"}, Default: "16:9"},
			{Key: "duration", Type: "select", Values: []string{"4", "5", "8", "10", "12", "15"}, Default: "5"},
			{Key: "resolution", Type: "select", Values: []string{"480p", "720p", "1080p"}, Default: "720p"},
			{Key: "turbo_mode", Type: "boolean", Values: []string{"false", "true"}, Default: "false"},
		}
	case "seedance-v2-video-extend":
		return []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"16:9", "9:16", "4:3", "3:4", "1:1", "21:9"}, Default: "16:9"},
			{Key: "duration", Type: "select", Values: []string{"4", "5", "8", "10", "12", "15"}, Default: "5"},
			{Key: "resolution", Type: "select", Values: []string{"720p", "1080p"}, Default: "720p"},
		}
	default:
		return nil
	}
}

func audioModelOptions(id string) []MediaOption {
	switch id {
	case "qwen3-tts":
		return []MediaOption{
			{Key: "language", Type: "select", Values: []string{
				"auto", "Chinese", "English", "German", "Italian", "Portuguese",
				"Spanish", "Japanese", "Korean", "French", "Russian",
			}, Default: "auto"},
			{Key: "voice", Type: "select", Values: []string{
				"Vivian", "Serena", "Ono_Anna", "Sohee",
				"Uncle_Fu", "Dylan", "Eric", "Ryan", "Aiden",
			}, Default: "Dylan"},
		}
	default:
		return nil
	}
}

func modelSupportsQuality(id string) bool {
	return id == "gpt-image-2" || id == "gpt-image-1.5"
}

func modelSupportsResolution(id string) bool {
	switch id {
	case "nano-banana-pro", "nano-banana-2", "gpt-image-2", "gpt-image-1.5":
		return true
	default:
		return false
	}
}

func normalizeImageResolution(modelID, resolution string) string {
	resolution = strings.ToLower(strings.TrimSpace(resolution))
	switch modelID {
	case "nano-banana-2":
		switch resolution {
		case "512px", "512", "0.5k":
			return "0.5k"
		case "1k", "2k", "4k":
			return resolution
		default:
			return resolution
		}
	default:
		switch resolution {
		case "1k", "2k", "4k":
			return resolution
		default:
			return strings.ToLower(resolution)
		}
	}
}

func qualityTierDefault(modelID string) string {
	if modelID == "kling-v3-pro" {
		return "pro"
	}
	return "std"
}

func normalizeImageOutputFormat(modelID, format string) string {
	format = strings.ToLower(strings.TrimSpace(format))
	if format == "" {
		return "png"
	}
	if modelID == "nano-banana-2" && format == "webp" {
		return "png"
	}
	return format
}
