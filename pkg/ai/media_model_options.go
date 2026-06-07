package ai

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
			{Key: "resolution", Type: "select", Values: []string{"512px", "1k", "2k", "4k"}, Default: "1k"},
			{Key: "output_format", Type: "select", Values: []string{"png", "jpeg", "webp"}, Default: "png"},
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
		return []MediaOption{
			{Key: "aspect_ratio", Type: "select", Values: []string{"16:9", "9:16", "1:1"}, Default: "16:9"},
			{Key: "duration", Type: "select", Values: []string{"5", "10"}, Default: "5"},
		}
	default:
		return nil
	}
}

func modelSupportsQuality(id string) bool {
	return id == "gpt-image-2" || id == "gpt-image-1.5"
}
