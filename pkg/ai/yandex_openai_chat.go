package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

const yandexOpenAIBase = "https://ai.api.cloud.yandex.net/v1"

// generateYandexOpenAIChat calls Yandex AI Studio OpenAI-compatible Chat Completions API.
// Required for Qwen and other models that do not support foundationModels completion.
func generateYandexOpenAIChat(
	ctx context.Context,
	cfg config.ConfigAI,
	messages []ChatMessage,
	modelSlug string,
	req TextRequest,
) (TextResponse, error) {
	oaiMessages := make([]map[string]any, 0, len(messages)+1)
	for _, m := range messages {
		oaiMessages = append(oaiMessages, map[string]any{
			"role":    m.Role,
			"content": m.Content,
		})
	}

	if strings.TrimSpace(req.ImageBase64) != "" {
		mime := strings.TrimSpace(req.ImageMimeType)
		if mime == "" {
			mime = "image/jpeg"
		}
		dataURL := "data:" + mime + ";base64," + strings.TrimSpace(req.ImageBase64)

		// Attach the image to the last user message (or add a new one).
		if len(oaiMessages) == 0 || oaiMessages[len(oaiMessages)-1]["role"] != "user" {
			oaiMessages = append(oaiMessages, map[string]any{
				"role": "user",
				"content": []map[string]any{
					{"type": "image_url", "image_url": map[string]any{"url": dataURL}},
				},
			})
		} else {
			prev := oaiMessages[len(oaiMessages)-1]
			text := ""
			if s, ok := prev["content"].(string); ok {
				text = s
			}
			prev["content"] = []map[string]any{
				{"type": "text", "text": text},
				{"type": "image_url", "image_url": map[string]any{"url": dataURL}},
			}
			oaiMessages[len(oaiMessages)-1] = prev
		}
	}

	maxTokens := cfg.TextMaxOutputTokens
	if req.MaxTokens != nil && *req.MaxTokens > 0 {
		maxTokens = *req.MaxTokens
	}

	payload := map[string]any{
		"model":      yandexModelURI(cfg.YandexFolderID, normalizeTextModelSlug(modelSlug)),
		"messages":   oaiMessages,
		"max_tokens": maxTokens,
	}
	if req.Temperature != nil {
		payload["temperature"] = *req.Temperature
	}

	headers := yandexHeaders(cfg.YandexAPIKey)
	headers["OpenAI-Project"] = cfg.YandexFolderID

	body, status, err := doJSON(ctx, cfg.HTTPTimeout, http.MethodPost, yandexOpenAIBase+"/chat/completions", headers, payload)
	if err != nil {
		return TextResponse{}, err
	}
	if status != http.StatusOK {
		return TextResponse{}, &ProviderError{Provider: "yandex-chat", Status: status, Message: truncate(string(body), 500)}
	}

	var resp struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(body, &resp); err != nil {
		return TextResponse{}, fmt.Errorf("yandex chat: decode: %w", err)
	}
	if len(resp.Choices) == 0 || resp.Choices[0].Message.Content == "" {
		return TextResponse{}, &ProviderError{Provider: "yandex-chat", Message: "empty completion"}
	}

	return finalizeTextResponse(resp.Choices[0].Message.Content, modelSlug), nil
}

func normalizeTextModelSlug(slug string) string {
	return strings.TrimSuffix(strings.TrimSpace(slug), "/latest")
}
