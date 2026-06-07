package ai

import (
	"context"
	"encoding/base64"
	"os"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func prepareWavespeedImageSource(ctx context.Context, cfg config.ConfigAI, req *ImageRequest) error {
	if strings.TrimSpace(req.SourceImageURL) != "" || strings.TrimSpace(req.ImageURL) != "" {
		return nil
	}

	b64 := strings.TrimSpace(req.ImageBase64)
	if b64 == "" {
		return nil
	}

	if i := strings.Index(b64, ","); i >= 0 {
		b64 = b64[i+1:]
	}

	data, err := base64.StdEncoding.DecodeString(b64)
	if err != nil {
		return &ProviderError{Provider: "wavespeed", Message: "invalid image base64"}
	}
	if len(data) == 0 {
		return &ProviderError{Provider: "wavespeed", Message: "image data is empty"}
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

func uploadWavespeedMediaBytes(
	ctx context.Context,
	cfg config.ConfigAI,
	data []byte,
	mimeType string,
	extFn func(string) string,
) (string, error) {
	if strings.TrimSpace(cfg.WavespeedAPIKey) == "" {
		return "", &ProviderError{Provider: "wavespeed", Message: "WAVESPEED_API_KEY is not configured"}
	}

	ext := extFn(mimeType)
	tmp, err := os.CreateTemp("", "wavespeed-upload-*"+ext)
	if err != nil {
		return "", err
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return "", err
	}
	if err := tmp.Close(); err != nil {
		return "", err
	}

	type uploadResult struct {
		url string
		err error
	}
	ch := make(chan uploadResult, 1)
	go func() {
		client := newWavespeedClient(cfg)
		url, err := client.Upload(tmpPath)
		ch <- uploadResult{url: url, err: err}
	}()

	select {
	case <-ctx.Done():
		return "", ctx.Err()
	case result := <-ch:
		if result.err != nil {
			return "", &ProviderError{Provider: "wavespeed", Message: truncate(result.err.Error(), 500)}
		}
		url := strings.TrimSpace(result.url)
		if url == "" {
			return "", &ProviderError{Provider: "wavespeed", Message: "upload completed without image url"}
		}
		return url, nil
	}
}

func extensionForImageMime(mimeType string) string {
	switch strings.ToLower(strings.TrimSpace(mimeType)) {
	case "image/jpeg", "image/jpg":
		return ".jpg"
	case "image/webp":
		return ".webp"
	case "image/gif":
		return ".gif"
	case "image/png":
		return ".png"
	default:
		return ".png"
	}
}
