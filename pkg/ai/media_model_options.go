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
	case "seedream-v4.5", "seedream-v5.0-lite":
		return []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"1:1", "16:9", "9:16", "4:3", "3:4"}, Default: "1:1"},
			{Key: "output_format", Type: "select", Values: []string{"jpeg", "png", "webp"}, Default: "jpeg"},
		}
	case "qwen-image", "qwen-image-2512":
		return []MediaOption{
			{Key: "size", Type: "select", Values: []string{"1024*1024", "1024x1024", "1328*1328"}, Default: "1024*1024"},
			{Key: "negative_prompt", Type: "text", Values: nil, Default: ""},
			{Key: "seed", Type: "number", Values: nil, Default: "-1"},
		}
	case "qwen-image-2.0":
		return []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"1:1", "16:9", "9:16", "4:3", "3:4", "3:2", "2:3"}, Default: "16:9"},
			{Key: "seed", Type: "number", Values: nil, Default: "-1"},
		}
	case "qwen-image-2.0-pro":
		return []MediaOption{
			{Key: "seed", Type: "number", Values: nil, Default: "-1"},
		}
	case "z-image-base", "z-image-turbo":
		return []MediaOption{
			{Key: "size", Type: "select", Values: []string{"1024*1024", "1024x1024"}, Default: "1024*1024"},
			{Key: "negative_prompt", Type: "text", Values: nil, Default: ""},
			{Key: "seed", Type: "number", Values: nil, Default: "-1"},
		}
	case "grok-imagine-edit":
		return []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"auto", "1:1", "16:9", "9:16", "4:3", "3:4", "3:2", "2:3"}, Default: "auto"},
			{Key: "resolution", Type: "select", Values: []string{"1k", "2k"}, Default: "1k"},
			{Key: "output_format", Type: "select", Values: []string{"jpeg", "png", "webp"}, Default: "jpeg"},
		}
	default:
		return nil
	}
}

func videoModelOptions(id string) []MediaOption {
	switch id {
	case "kling-v3-std", "kling-v3-pro", "kling-v3-4k":
		opts := []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"16:9", "9:16", "1:1"}, Default: "16:9"},
			{Key: "duration", Type: "select", Values: klingDurationOptionValues(), Default: "5"},
			{Key: "resolution", Type: "select", Values: []string{"720p", "1080p", "4k"}, Default: klingResolutionDefault(id)},
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
	case "wan-2.5-t2v", "wan-2.7-t2v":
		return []MediaOption{
			{Key: "duration", Type: "select", Values: []string{"2", "5", "10", "15"}, Default: "5"},
			{Key: "resolution", Type: "select", Values: []string{"480P", "720P", "1080P"}, Default: "720P"},
			{Key: "negative_prompt", Type: "text", Values: nil, Default: ""},
		}
	case "wan-2.6-i2v", "wan-2.2-spicy-i2v":
		return []MediaOption{
			{Key: "duration", Type: "select", Values: []string{"5", "8"}, Default: "5"},
			{Key: "resolution", Type: "select", Values: []string{"480P", "720P", "1080p"}, Default: "720P"},
		}
	case "wan-2.7-flf", "wan-2.7-grid":
		return []MediaOption{
			{Key: "duration", Type: "select", Values: []string{"5", "10"}, Default: "5"},
		}
	case "happyhorse-t2v", "happyhorse-i2v", "happyhorse-ref2v":
		return []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"16:9", "9:16", "1:1", "4:3", "3:4"}, Default: "16:9"},
			{Key: "duration", Type: "select", Values: []string{"3", "5", "10", "15"}, Default: "5"},
			{Key: "resolution", Type: "select", Values: []string{"720p", "1080p"}, Default: "720p"},
		}
	case "happyhorse-video-extend":
		return []MediaOption{
			{Key: "duration", Type: "select", Values: []string{"3", "5", "10"}, Default: "5"},
			{Key: "extend_by", Type: "select", Values: []string{"3", "5", "10"}, Default: "5"},
		}
	case "sora-2-t2v", "sora-2-i2v", "sora-2-t2v-pro":
		return []MediaOption{
			{Key: "duration", Type: "select", Values: []string{"5", "10"}, Default: "5"},
			{Key: "resolution", Type: "select", Values: []string{"720p", "1080p"}, Default: "720p"},
		}
	case "veo-3.1-extend":
		return []MediaOption{
			{Key: "resolution", Type: "select", Values: []string{"720p", "1080p"}, Default: "1080p"},
			{Key: "negative_prompt", Type: "text", Values: nil, Default: ""},
		}
	case "vidu-q3-i2v-spicy":
		return []MediaOption{
			{Key: "duration", Type: "select", Values: []string{"1", "5", "10", "16"}, Default: "5"},
			{Key: "resolution", Type: "select", Values: []string{"540p", "720p", "1080p"}, Default: "720p"},
			{Key: "generate_audio", Type: "boolean", Values: []string{"true", "false"}, Default: "true"},
			{Key: "movement_amplitude", Type: "select", Values: []string{"auto", "small", "medium", "large"}, Default: "auto"},
		}
	case "hailuo-2.3-t2v", "hailuo-2.3-i2v-fast", "hailuo-2.3-i2v-pro":
		return []MediaOption{
			{Key: "duration", Type: "select", Values: []string{"6"}, Default: "6"},
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
	case "omnivoice":
		return []MediaOption{
			{Key: "speed", Type: "select", Values: []string{"0.8", "1.0", "1.2", "1.5"}, Default: "1.0"},
		}
	case "elevenlabs-v3":
		return []MediaOption{
			{Key: "voice", Type: "select", Values: []string{
				"Alice", "Aria", "Roger", "Sarah", "Laura", "Charlie", "George",
				"Callum", "River", "Liam", "Charlotte", "Matilda", "Will", "Jessica",
				"Eric", "Chris", "Brian", "Daniel", "Lily", "Bill",
			}, Default: "Alice"},
		}
	case "minimax-speech-2.6":
		return []MediaOption{
			{Key: "voice", Type: "select", Values: []string{
				"Wise_Woman", "Friendly_Person", "Inspirational_girl", "Deep_Voice_Man",
				"Calm_Woman", "Casual_Guy", "Lively_Girl", "Patient_Man", "Young_Knight",
				"Determined_Man", "Lovely_Girl", "Decent_Boy", "Imposing_Manner", "Elegant_Man",
				"Abbess", "Sweet_Girl_2", "Exuberant_Girl",
			}, Default: "Friendly_Person"},
			{Key: "emotion", Type: "select", Values: []string{
				"happy", "sad", "angry", "fearful", "disgusted", "surprised", "neutral",
			}, Default: "happy"},
		}
	case "mureka-v9":
		return []MediaOption{
			{Key: "number_of_songs", Type: "select", Values: []string{"1", "2", "3"}, Default: "1"},
			{Key: "output_format", Type: "select", Values: []string{"mp3", "wav", "flac"}, Default: "mp3"},
		}
	case "ace-step-1.5":
		return []MediaOption{
			{Key: "duration", Type: "select", Values: []string{"30", "60", "120", "180", "240"}, Default: "60"},
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
	case "nano-banana-pro", "nano-banana-2", "gpt-image-2", "gpt-image-1.5", "grok-imagine-edit":
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

func klingResolutionDefault(modelID string) string {
	switch modelID {
	case "kling-v3-4k":
		return "4k"
	case "kling-v3-pro":
		return "1080p"
	default:
		return "720p"
	}
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
