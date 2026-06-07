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
	AudioModels []MediaModel `json:"audio_models,omitempty"`
}

// ImageRequest is the body for POST /v1/generate/image.
type ImageRequest struct {
	Prompt         string        `json:"prompt"`
	Text           string        `json:"text"`
	Model          string        `json:"model"`
	Size           string        `json:"size"`
	AspectRatio    string        `json:"aspect_ratio"`
	Resolution     string        `json:"resolution"`
	Quality        string        `json:"quality"`
	OutputFormat   string        `json:"output_format"`
	Mode           string        `json:"mode"` // generate | edit | multi
	NumImages      int           `json:"num_images"`
	SourceImageURL string        `json:"sourceImageUrl"`
	ImageURL       string        `json:"image_url"`
	ImageBase64    string        `json:"imageBase64"`
	ImageMimeType  string        `json:"imageMimeType"`
	Messages       []ChatMessage `json:"messages"`
}

// ImageResponse is returned by POST /v1/generate/image.
type ImageResponse struct {
	ImageURL    string   `json:"image_url,omitempty"`
	ImageURLs   []string `json:"image_urls,omitempty"`
	ImageBase64 string   `json:"image_base64,omitempty"`
	Model       string   `json:"model"`
}

// VideoRequest is the body for POST /v1/generate/video.
type VideoRequest struct {
	Prompt         string        `json:"prompt"`
	Text           string        `json:"text"`
	Model          string        `json:"model"`
	AspectRatio    string        `json:"aspect_ratio"`
	Duration       int           `json:"duration"`
	Resolution     string        `json:"resolution"`
	Quality        string        `json:"quality"`
	SourceImageURL string        `json:"sourceImageUrl"`
	SourceVideoURL string        `json:"sourceVideoUrl"`
	ImageURL       string        `json:"image_url"`
	VideoURL       string        `json:"video_url"`
	LastImageURL    string        `json:"last_image"`
	ReferenceImages []string      `json:"reference_images"`
	GenerateAudio   *bool         `json:"generate_audio"`
	CameraFixed     *bool         `json:"camera_fixed"`
	TurboMode       *bool         `json:"turbo_mode"`
	Seed            int           `json:"seed"`
	Messages        []ChatMessage `json:"messages"`
}

// VideoResponse is returned by POST /v1/generate/video.
type VideoResponse struct {
	VideoURL string `json:"video_url"`
	Model    string `json:"model"`
}

// AudioRequest is the body for POST /v1/generate/audio.
type AudioRequest struct {
	Prompt           string        `json:"prompt"`
	Text             string        `json:"text"`
	Model            string        `json:"model"`
	Language         string        `json:"language"`
	Voice            string        `json:"voice"`
	StyleInstruction string        `json:"style_instruction"`
	ReferenceText    string        `json:"reference_text"`
	Mode             string        `json:"mode"` // tts | clone
	SourceAudioURL   string        `json:"sourceAudioUrl"`
	AudioURL         string        `json:"audio_url"`
	AudioBase64      string        `json:"audioBase64"`
	AudioMimeType    string        `json:"audioMimeType"`
	Messages         []ChatMessage `json:"messages"`
}

// AudioResponse is returned by POST /v1/generate/audio.
type AudioResponse struct {
	AudioURL string `json:"audio_url"`
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
		AudioModels: ListAudioModels(),
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

func (s *Service) GenerateAudio(ctx context.Context, req AudioRequest) (AudioResponse, error) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(req.Text)
	}
	if prompt == "" {
		return AudioResponse{}, ErrPromptEmpty
	}

	if !s.cfg.WavespeedImageEnabled() {
		return AudioResponse{}, &ProviderError{
			Provider: "wavespeed",
			Message:  "WAVESPEED_API_KEY is not configured",
		}
	}

	def, ok := resolveWavespeedAudioModel(req.Model)
	if !ok {
		def, ok = resolveWavespeedAudioModel("qwen3-tts")
	}
	if !ok {
		return AudioResponse{}, &ProviderError{Provider: "wavespeed", Message: "unknown audio model: " + req.Model}
	}

	if hasQwen3TTSSourceAudio(req) && strings.TrimSpace(req.SourceAudioURL) == "" && strings.TrimSpace(req.AudioURL) == "" && strings.TrimSpace(req.AudioBase64) == "" {
		return AudioResponse{}, &ProviderError{Provider: "wavespeed", Message: "reference audio is required for voice clone"}
	}

	return generateWavespeedAudio(ctx, s.cfg, prompt, req, def)
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
