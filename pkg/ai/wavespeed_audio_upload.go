package ai

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func prepareWavespeedAudioSource(ctx context.Context, cfg config.ConfigAI, req *AudioRequest) error {
	if strings.TrimSpace(req.SourceAudioURL) != "" || strings.TrimSpace(req.AudioURL) != "" {
		return nil
	}

	b64 := strings.TrimSpace(req.AudioBase64)
	if b64 == "" {
		return nil
	}

	if i := strings.Index(b64, ","); i >= 0 {
		b64 = b64[i+1:]
	}

	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return &ProviderError{Provider: "wavespeed", Message: "invalid audio base64"}
	}
	if len(data) == 0 {
		return &ProviderError{Provider: "wavespeed", Message: "audio data is empty"}
	}
	if len(data) > 8<<20 {
		return &ProviderError{Provider: "wavespeed", Message: "audio is too large (max 8 MB)"}
	}

	url, err := uploadWavespeedMediaBytes(ctx, cfg, data, req.AudioMimeType, extensionForAudioMime)
	if err != nil {
		return err
	}
	req.SourceAudioURL = url
	return nil
}

func extensionForAudioMime(mimeType string) string {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "audio/wav", "audio/x-wav":
		return ".wav"
	case "audio/mpeg", "audio/mp3":
		return ".mp3"
	case "audio/ogg":
		return ".ogg"
	case "audio/webm":
		return ".webm"
	case "audio/mp4", "audio/m4a", "audio/x-m4a":
		return ".m4a"
	default:
		return ".wav"
	}
}
