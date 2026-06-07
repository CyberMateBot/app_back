package ai

import (
	"context"
	"fmt"
	"strings"

	wsapi "github.com/WaveSpeedAI/wavespeed-go/api"
	"github.com/twelvepills-936/tgapp-/pkg/config"
)

func newWavespeedClient(cfg config.ConfigAI) *wsapi.Client {
	return wsapi.NewClient(wsapi.WithAPIKey(cfg.WavespeedAPIKey))
}

func runWavespeedModel(ctx context.Context, cfg config.ConfigAI, modelSlug string, input map[string]any) ([]string, error) {
	if strings.TrimSpace(cfg.WavespeedAPIKey) == "" {
		return nil, &ProviderError{Provider: "wavespeed", Message: "WAVESPEED_API_KEY is not configured"}
	}

	client := newWavespeedClient(cfg)
	timeoutSec := cfg.ImagePollTimeout.Seconds()
	if timeoutSec <= 0 {
		timeoutSec = 180
	}
	pollSec := float64(cfg.NanoBananaPollMS) / 1000.0
	if pollSec <= 0 {
		pollSec = 2
	}

	type runResult struct {
		out map[string]any
		err error
	}
	ch := make(chan runResult, 1)
	go func() {
		out, err := client.Run(modelSlug, input,
			wsapi.WithTimeout(timeoutSec),
			wsapi.WithPollInterval(pollSec),
			wsapi.WithSyncMode(cfg.NanoBananaSyncMode),
		)
		ch <- runResult{out: out, err: err}
	}()

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case result := <-ch:
		if result.err != nil {
			return nil, &ProviderError{Provider: "wavespeed", Message: truncate(result.err.Error(), 500)}
		}
		urls, err := extractWavespeedOutputs(result.out)
		if err != nil {
			return nil, err
		}
		return urls, nil
	}
}

func extractWavespeedOutputs(out map[string]any) ([]string, error) {
	raw, ok := out["outputs"]
	if !ok || raw == nil {
		return nil, &ProviderError{Provider: "wavespeed", Message: "completed without output url"}
	}

	switch items := raw.(type) {
	case []any:
		urls := make([]string, 0, len(items))
		for _, item := range items {
			if s, ok := item.(string); ok && strings.TrimSpace(s) != "" {
				urls = append(urls, strings.TrimSpace(s))
			}
		}
		if len(urls) == 0 {
			return nil, &ProviderError{Provider: "wavespeed", Message: "completed without output url"}
		}
		return urls, nil
	case []string:
		if len(items) == 0 {
			return nil, &ProviderError{Provider: "wavespeed", Message: "completed without output url"}
		}
		return items, nil
	default:
		return nil, fmt.Errorf("wavespeed: unexpected outputs type %T", raw)
	}
}
