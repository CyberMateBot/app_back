package ai

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func prepareWavespeedVideoSource(ctx context.Context, cfg config.ConfigAI, req *VideoRequest) error {
	if err := prepareWavespeedVideoImageSource(ctx, cfg, req); err != nil {
		return err
	}
	if err := prepareWavespeedVideoFrameSources(ctx, cfg, req); err != nil {
		return err
	}
	return prepareWavespeedVideoFileSource(ctx, cfg, req)
}

func prepareWavespeedVideoImageSource(ctx context.Context, cfg config.ConfigAI, req *VideoRequest) error {
	if strings.TrimSpace(req.SourceImageURL) != "" || strings.TrimSpace(req.ImageURL) != "" {
		return nil
	}

	b64 := strings.TrimSpace(req.ImageBase64)
	if b64 == "" {
		return nil
	}

	data, err := decodeMediaBase64(b64)
	if err != nil {
		return err
	}
	if len(data) > 8<<20 {
		return &ProviderError{Provider: "wavespeed", Message: "image is too large (max 8 MB)"}
	}

	url, err := uploadWavespeedMediaBytes(ctx, cfg, data, req.ImageMimeType, extensionForImageMime)
	if err != nil {
		return err
	}
	req.SourceImageURL = url
	return nil
}

func prepareWavespeedVideoFileSource(ctx context.Context, cfg config.ConfigAI, req *VideoRequest) error {
	if strings.TrimSpace(req.SourceVideoURL) != "" || strings.TrimSpace(req.VideoURL) != "" {
		return nil
	}

	b64 := strings.TrimSpace(req.VideoBase64)
	if b64 == "" {
		return nil
	}

	data, err := decodeMediaBase64(b64)
	if err != nil {
		return err
	}
	if len(data) > 16<<20 {
		return &ProviderError{Provider: "wavespeed", Message: "video is too large (max 16 MB)"}
	}

	url, err := uploadWavespeedMediaBytes(ctx, cfg, data, req.VideoMimeType, extensionForVideoMime)
	if err != nil {
		return err
	}
	req.SourceVideoURL = url
	return nil
}

func prepareWavespeedVideoFrameSources(ctx context.Context, cfg config.ConfigAI, req *VideoRequest) error {
	if strings.TrimSpace(req.FirstFrameURL) == "" {
		if b64 := strings.TrimSpace(req.FirstFrameBase64); b64 != "" {
			url, err := uploadDecodedImage(ctx, cfg, b64, req.FirstFrameMimeType)
			if err != nil {
				return err
			}
			req.FirstFrameURL = url
		}
	}
	if strings.TrimSpace(req.LastFrameURL) == "" {
		if b64 := strings.TrimSpace(req.LastFrameBase64); b64 != "" {
			url, err := uploadDecodedImage(ctx, cfg, b64, req.LastFrameMimeType)
			if err != nil {
				return err
			}
			req.LastFrameURL = url
		}
	}
	return nil
}

func uploadDecodedImage(ctx context.Context, cfg config.ConfigAI, b64, mimeType string) (string, error) {
	data, err := decodeMediaBase64(b64)
	if err != nil {
		return "", err
	}
	if len(data) > 8<<20 {
		return "", &ProviderError{Provider: "wavespeed", Message: "image is too large (max 8 MB)"}
	}
	return uploadWavespeedMediaBytes(ctx, cfg, data, mimeType, extensionForImageMime)
}

func decodeMediaBase64(b64 string) ([]byte, error) {
	if i := strings.Index(b64, ","); i >= 0 {
		b64 = b64[i+1:]
	}
	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return nil, &ProviderError{Provider: "wavespeed", Message: "invalid media base64"}
	}
	if len(data) == 0 {
		return nil, &ProviderError{Provider: "wavespeed", Message: "media data is empty"}
	}
	return data, nil
}

func extensionForVideoMime(mimeType string) string {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "video/webm":
		return ".webm"
	case "video/quicktime":
		return ".mov"
	default:
		return ".mp4"
	}
}
