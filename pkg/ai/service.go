package ai

import (
	"context"
	"strings"

	"github.com/twelvepills-936/tgapp-/pkg/config"
)

// TextRequest is the body for POST /v1/generate/text.
type TextRequest struct {
	Prompt        string        `json:"prompt"`
	Text          string        `json:"text"` // alias for prompt
	Model         string        `json:"model"`
	System        string        `json:"system"`
	Messages      []ChatMessage `json:"messages"`
	ImageBase64   string        `json:"imageBase64"`
	ImageMimeType string        `json:"imageMimeType"`
	Temperature   *float64      `json:"temperature"`
	MaxTokens     *int          `json:"max_tokens"`
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
	Text    string `json:"text"` // some clients send "text" instead of "content"
}

// TextResponse is returned by POST /v1/generate/text.
type TextResponse struct {
	Text   string `json:"text"`
	Model  string `json:"model"`
	Format string `json:"format"` // markdown — render with Markdown + LaTeX ($...$) on frontend
}

// ModelsResponse is returned by GET /v1/generate/models.
type ModelsResponse struct {
	TextModels  []TextModel  `json:"text_models"`
	ImageModels []MediaModel `json:"image_models,omitempty"`
	VideoModels []MediaModel `json:"video_models,omitempty"`
}

// ImageRequest is the body for POST /v1/generate/image.
type ImageRequest struct {
	Prompt      string `json:"prompt"`
	Text        string `json:"text"`
	Model       string `json:"model"`
	Size        string `json:"size"`
	AspectRatio string `json:"aspect_ratio"`
}

// ImageResponse is returned by POST /v1/generate/image.
type ImageResponse struct {
	ImageURL    string `json:"image_url,omitempty"`
	ImageBase64 string `json:"image_base64,omitempty"`
	Model       string `json:"model"`
}

// VideoRequest is the body for POST /v1/generate/video.
type VideoRequest struct {
	Prompt      string `json:"prompt"`
	Text        string `json:"text"`
	Model       string `json:"model"`
	AspectRatio string `json:"aspect_ratio"`
	Duration    int    `json:"duration"`
}

// VideoResponse is returned by POST /v1/generate/video.
type VideoResponse struct {
	VideoURL string `json:"video_url"`
	Model    string `json:"model"`
}

// Service routes generation to configured providers.
type Service struct {
	cfg config.ConfigAI
}

func NewService(cfg config.ConfigAI) *Service {
	return &Service{cfg: cfg}
}

func (s *Service) ListModels() ModelsResponse {
	return ModelsResponse{
		TextModels:  ListTextModels(),
		ImageModels: ListImageModels(),
		VideoModels: ListVideoModels(),
	}
}

func (s *Service) GenerateText(ctx context.Context, req TextRequest) (TextResponse, error) {
	prompt, messages, err := normalizeTextInput(req)
	if err != nil {
		return TextResponse{}, err
	}

	provider := strings.ToLower(strings.TrimSpace(req.Model))
	switch provider {
	case "openai", "gpt":
		if s.cfg.OpenAITextEnabled() {
			return generateOpenAICompatText(ctx, s.cfg, "https://api.openai.com/v1", s.cfg.OpenAITextModel, messages, prompt, req)
		}
	}

	def, ok := resolveTextModel(req.Model, s.cfg.YandexGPTModel)
	if ok {
		if def.UseWavespeed {
			if !s.cfg.WavespeedTextEnabled() {
				return TextResponse{}, &ProviderError{
					Provider: "wavespeed",
					Message:  "WAVESPEED_API_KEY is not configured",
				}
			}
			slug := def.Slug
			if slug == "" {
				slug = s.cfg.GeminiModel
			}
			out, err := generateOpenAICompatText(ctx, s.cfg, s.cfg.GeminiAPIBaseURL, slug, messages, prompt, req)
			if err != nil {
				return out, err
			}
			out.Model = def.ID
			return out, nil
		}

		if s.cfg.YandexTextEnabled() {
			if def.UseOpenAIChat {
				return generateYandexOpenAIChat(ctx, s.cfg, messages, def.Slug, req)
			}
			if def.UseResponses {
				return generateYandexResponsesText(ctx, s.cfg, messages, def.Slug, req)
			}
			slug := def.Slug
			if slug == "" {
				slug = s.cfg.YandexGPTModel
			}
			return generateYandexText(ctx, s.cfg, messages, prompt, slug, req)
		}
	}

	if s.cfg.YandexTextEnabled() {
		return generateYandexText(ctx, s.cfg, messages, prompt, req.Model, req)
	}
	if s.cfg.OpenAITextEnabled() {
		return generateOpenAICompatText(ctx, s.cfg, "https://api.openai.com/v1", s.cfg.OpenAITextModel, messages, prompt, req)
	}
	if s.cfg.WavespeedTextEnabled() {
		return generateOpenAICompatText(ctx, s.cfg, s.cfg.GeminiAPIBaseURL, s.cfg.GeminiModel, messages, prompt, req)
	}

	return TextResponse{}, ErrNotConfigured
}

func (s *Service) GenerateImage(ctx context.Context, req ImageRequest) (ImageResponse, error) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(req.Text)
	}
	if prompt == "" {
		return ImageResponse{}, ErrPromptEmpty
	}

	if def, ok := resolveWavespeedImageModel(req.Model); ok {
		if !s.cfg.WavespeedImageEnabled() {
			return ImageResponse{}, &ProviderError{
				Provider: "wavespeed",
				Message:  "WAVESPEED_API_KEY is not configured",
			}
		}
		return generateWavespeedImage(ctx, s.cfg, prompt, req, def)
	}

	provider := strings.ToLower(strings.TrimSpace(req.Model))
	switch provider {
	case "alice-ai-art":
		if s.cfg.YandexTextEnabled() {
			return generateYandexOpenAIImage(ctx, s.cfg, prompt, req)
		}
		return ImageResponse{}, ErrNotConfigured
	case "", "yandex", "alice", "yandex-art", "yandex-art-2.0", "default":
		if s.cfg.YandexTextEnabled() {
			return generateYandexImage(ctx, s.cfg, prompt, req)
		}
	default:
		if s.cfg.YandexTextEnabled() {
			return generateYandexImage(ctx, s.cfg, prompt, req)
		}
	}

	if s.cfg.YandexTextEnabled() {
		return generateYandexImage(ctx, s.cfg, prompt, req)
	}
	if s.cfg.WavespeedImageEnabled() {
		def, _ := resolveWavespeedImageModel("nano-banana")
		return generateWavespeedImage(ctx, s.cfg, prompt, req, def)
	}

	return ImageResponse{}, ErrNotConfigured
}

func (s *Service) GenerateVideo(ctx context.Context, req VideoRequest) (VideoResponse, error) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(req.Text)
	}
	if prompt == "" {
		return VideoResponse{}, ErrPromptEmpty
	}

	if !s.cfg.WavespeedImageEnabled() {
		return VideoResponse{}, &ProviderError{
			Provider: "wavespeed",
			Message:  "WAVESPEED_API_KEY is not configured",
		}
	}

	def, ok := resolveWavespeedVideoModel(req.Model)
	if !ok {
		def, ok = resolveWavespeedVideoModel("kling-v3-std")
	}
	if !ok {
		return VideoResponse{}, &ProviderError{Provider: "wavespeed", Message: "unknown video model: " + req.Model}
	}

	return generateWavespeedVideo(ctx, s.cfg, prompt, req, def)
}

func normalizeTextInput(req TextRequest) (prompt string, messages []ChatMessage, err error) {
	prompt = strings.TrimSpace(req.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(req.Text)
	}

	messages = req.Messages
	if len(messages) == 0 && prompt != "" {
		messages = []ChatMessage{{Role: "user", Content: prompt}}
	}
	if len(messages) == 0 {
		return "", nil, ErrPromptEmpty
	}

	if strings.TrimSpace(req.System) != "" {
		messages = append([]ChatMessage{{Role: "system", Content: req.System}}, messages...)
	}

	if len(messages) > 0 && messages[0].Role == "system" {
		messages[0].Content = mergeInstructions(messages[0].Content)
	} else {
		messages = append([]ChatMessage{{Role: "system", Content: mergeInstructions("")}}, messages...)
	}

	for i := range messages {
		if messages[i].Content == "" {
			messages[i].Content = messages[i].Text
		}
		if messages[i].Role == "" {
			messages[i].Role = "user"
		}
	}
	return prompt, messages, nil
}
