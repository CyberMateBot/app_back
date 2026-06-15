package ai

import (
	"context"
	"encoding/base64"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func prepareWavespeedThreeDSource(ctx context.Context, cfg config.ConfigAI, req *ThreeDRequest) error {
	if err := prepareSingleThreeDImage(ctx, cfg, req); err != nil {
		return err
	}
	return prepareMultiThreeDImages(ctx, cfg, req)
}

func prepareSingleThreeDImage(ctx context.Context, cfg config.ConfigAI, req *ThreeDRequest) error {
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

	url, err := uploadWavespeedMediaBytes(ctx, cfg, data, req.ImageMimeType, extensionForImageMime)
	if err != nil {
		return err
	}
	req.SourceImageURL = url
	return nil
}

func prepareMultiThreeDImages(ctx context.Context, cfg config.ConfigAI, req *ThreeDRequest) error {
	if len(req.SourceImages) > 0 {
		return nil
	}
	if len(req.Images) > 0 {
		req.SourceImages = append([]string(nil), req.Images...)
		return nil
	}
	if len(req.ImageBase64List) == 0 {
		return nil
	}

	uploaded := make([]string, 0, len(req.ImageBase64List))
	for i, entry := range req.ImageBase64List {
		b64 := strings.TrimSpace(entry)
		if b64 == "" {
			continue
		}
		if j := strings.Index(b64, ","); j >= 0 {
			b64 = b64[j+1:]
		}
		data, err := base64.StdEncoding.DecodeString(b64)
		if err != nil {
			return &ProviderError{Provider: "wavespeed", Message: "invalid multiview image base64"}
		}
		mime := ""
		if i < len(req.ImageMimeTypes) {
			mime = req.ImageMimeTypes[i]
		}
		url, err := uploadWavespeedMediaBytes(ctx, cfg, data, mime, extensionForImageMime)
		if err != nil {
			return err
		}
		uploaded = append(uploaded, url)
	}
	if len(uploaded) > 0 {
		req.SourceImages = uploaded
	}
	return nil
}
