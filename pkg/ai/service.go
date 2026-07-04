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
	TextModels   []TextModel  `json:"text_models"`
	ImageModels  []MediaModel `json:"image_models,omitempty"`
	VideoModels  []MediaModel `json:"video_models,omitempty"`
	AudioModels  []MediaModel `json:"audio_models,omitempty"`
	ThreeDModels []MediaModel `json:"three_d_models,omitempty"`
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
	Mode           string        `json:"mode"` // generate | edit | multi | sequential | edit-sequential
	NumImages      int           `json:"num_images"`
	NegativePrompt string        `json:"negative_prompt"`
	Seed           int           `json:"seed"`
	PromptExtend   *bool         `json:"prompt_extend"`
	Strength       float64       `json:"strength"`
	GuidanceScale  float64       `json:"guidance_scale"`
	NumInferenceSteps int        `json:"num_inference_steps"`
	Width          int           `json:"width"`
	Height         int           `json:"height"`
	SourceImageURL string        `json:"sourceImageUrl"`
	ImageURL       string        `json:"image_url"`
	ImageBase64    string        `json:"imageBase64"`
	ImageMimeType  string        `json:"imageMimeType"`
	WebSearch      bool          `json:"web_search"`
	ImageSearch    bool          `json:"image_search"`
	Messages       []ChatMessage `json:"messages"`
	TelegramID     string        `json:"telegramId"`
	InitDataRaw    string        `json:"initDataRaw"`
	SessionID      string        `json:"sessionId"`
	Category       string        `json:"category"`
}

// ImageResponse is returned by POST /v1/generate/image.
type ImageResponse struct {
	ImageURL    string   `json:"image_url,omitempty"`
	ImageURLs   []string `json:"image_urls,omitempty"`
	ImageBase64 string   `json:"image_base64,omitempty"`
	Model       string   `json:"model"`
}

// CameraControlConfig holds per-axis camera movement for Kling simple mode.
type CameraControlConfig struct {
	Horizontal float64 `json:"horizontal,omitempty"`
	Vertical   float64 `json:"vertical,omitempty"`
	Pan        float64 `json:"pan,omitempty"`
	Tilt       float64 `json:"tilt,omitempty"`
	Roll       float64 `json:"roll,omitempty"`
	Zoom       float64 `json:"zoom,omitempty"`
}

// CameraControl configures optional Kling camera motion.
type CameraControl struct {
	Type   string               `json:"type,omitempty"`
	Config *CameraControlConfig `json:"config,omitempty"`
}

// VideoRequest is the body for POST /v1/generate/video.
type VideoRequest struct {
	Prompt            string         `json:"prompt"`
	Text              string         `json:"text"`
	Model             string         `json:"model"`
	AspectRatio       string         `json:"aspect_ratio"`
	Duration          int            `json:"duration"`
	Resolution        string         `json:"resolution"`
	Quality           string         `json:"quality"`
	NegativePrompt    string         `json:"negative_prompt"`
	SourceImageURL    string         `json:"sourceImageUrl"`
	SourceVideoURL    string         `json:"sourceVideoUrl"`
	ImageURL          string         `json:"image_url"`
	VideoURL          string         `json:"video_url"`
	ImageBase64       string         `json:"imageBase64"`
	ImageMimeType     string         `json:"imageMimeType"`
	VideoBase64       string         `json:"videoBase64"`
	VideoMimeType     string         `json:"videoMimeType"`
	LastImageURL      string         `json:"last_image"`
	ReferenceImages   []string       `json:"reference_images"`
	GenerateAudio     *bool          `json:"generate_audio"`
	Sound             *bool          `json:"sound"`
	CameraControl     *CameraControl `json:"camera_control"`
	CameraFixed       *bool          `json:"camera_fixed"`
	TurboMode         *bool          `json:"turbo_mode"`
	Seed              int            `json:"seed"`
	ExtendBy          int            `json:"extend_by"`
	EnablePromptExpansion *bool        `json:"enable_prompt_expansion"`
	GoFast            *bool          `json:"go_fast"`
	EditInstruction   string         `json:"edit_instruction"`
	FirstFrameURL     string         `json:"first_frame_url"`
	LastFrameURL      string         `json:"last_frame_url"`
	FirstFrameBase64  string         `json:"firstFrameBase64"`
	FirstFrameMimeType string        `json:"firstFrameMimeType"`
	LastFrameBase64   string         `json:"lastFrameBase64"`
	LastFrameMimeType string         `json:"lastFrameMimeType"`
	ImageGridURL      string         `json:"image_grid_url"`
	BGM               *bool          `json:"bgm"`
	MovementAmplitude string         `json:"movement_amplitude"`
	Messages          []ChatMessage  `json:"messages"`
	TelegramID        string         `json:"telegramId"`
	InitDataRaw       string         `json:"initDataRaw"`
	SessionID         string         `json:"sessionId"`
	Category          string         `json:"category"`
}

// VideoResponse is returned by POST /v1/generate/video.
type VideoResponse struct {
	VideoURL string `json:"video_url"`
	Model    string `json:"model"`
}

// AudioRequest is the body for POST /v1/generate/audio.
type AudioRequest struct {
	Prompt               string        `json:"prompt"`
	Text                 string        `json:"text"`
	Model                string        `json:"model"`
	Language             string        `json:"language"`
	Voice                string        `json:"voice"`
	StyleInstruction     string        `json:"style_instruction"`
	ReferenceText        string        `json:"reference_text"`
	Mode                 string        `json:"mode"` // tts | clone
	SourceAudioURL       string        `json:"sourceAudioUrl"`
	AudioURL             string        `json:"audio_url"`
	AudioBase64          string        `json:"audioBase64"`
	AudioMimeType        string        `json:"audioMimeType"`
	Speed                float64       `json:"speed"`
	Pitch                int           `json:"pitch"`
	Volume               float64       `json:"volume"`
	Emotion              string        `json:"emotion"`
	Similarity           float64       `json:"similarity"`
	Stability            float64       `json:"stability"`
	UseSpeakerBoost      *bool         `json:"use_speaker_boost"`
	Tags                 string        `json:"tags"`
	Duration             int           `json:"duration"`
	Seed                 int           `json:"seed"`
	NumberOfSongs        int           `json:"number_of_songs"`
	OutputFormat         string        `json:"output_format"`
	LanguageBoost        string        `json:"language_boost"`
	EnglishNormalization *bool         `json:"english_normalization"`
	Format               string        `json:"format"`
	Messages             []ChatMessage `json:"messages"`
	TelegramID           string        `json:"telegramId"`
	InitDataRaw          string        `json:"initDataRaw"`
}

// AudioResponse is returned by POST /v1/generate/audio.
type AudioResponse struct {
	AudioURL string `json:"audio_url"`
	Model    string `json:"model"`
}

// ThreeDRequest is the body for POST /v1/generate/3d.
type ThreeDRequest struct {
	Prompt                string   `json:"prompt"`
	Text                  string   `json:"text"`
	Model                 string   `json:"model"`
	NegativePrompt        string   `json:"negative_prompt"`
	SourceImageURL        string   `json:"sourceImageUrl"`
	ImageURL              string   `json:"image_url"`
	ImageBase64           string   `json:"imageBase64"`
	ImageMimeType         string   `json:"imageMimeType"`
	SourceImages          []string `json:"sourceImages"`
	Images                []string `json:"images"`
	ImageBase64List       []string `json:"imageBase64List"`
	ImageMimeTypes        []string `json:"imageMimeTypes"`
	TextureQuality        string   `json:"texture_quality"`
	FaceLimit             int      `json:"face_limit"`
	Quad                  *bool    `json:"quad"`
	PBR                   *bool    `json:"pbr"`
	OutputFormat          string   `json:"output_format"`
	Texture               *bool    `json:"texture"`
	GeometryQuality       string   `json:"geometry_quality"`
	AutoSize              *bool    `json:"auto_size"`
	Mode                  string   `json:"mode"`
	ArtStyle              string   `json:"art_style"`
	Topology              string   `json:"topology"`
	TargetPolycount       int      `json:"target_polycount"`
	EnablePBR             *bool    `json:"enable_pbr"`
	EnableGeometry        *bool    `json:"enable_geometry"`
	EnablePromptExpansion *bool    `json:"enable_prompt_expansion"`
	TAPose                *bool    `json:"ta_pose"`
	SymmetryMode          string   `json:"symmetry_mode"`
	ShouldRemesh          *bool    `json:"should_remesh"`
	GenerateType          string   `json:"generate_type"`
	Tier                  string   `json:"tier"`
	QualityAndMesh        string   `json:"quality_and_mesh"`
	Material              string   `json:"material"`
	HDTexture             *bool    `json:"hd_texture"`
	Addons                string   `json:"addons"`
	GeometryFileFormat    string   `json:"geometry_file_format"`
	TextureMode           string   `json:"texture_mode"`
	GeometryInstructMode  string   `json:"geometry_instruct_mode"`
	TextureDelight        *bool    `json:"texture_delight"`
	IsSymmetric           string   `json:"is_symmetric"`
	IsMicro               *bool    `json:"is_micro"`
	PreviewRender         *bool    `json:"preview_render"`
	TelegramID            string   `json:"telegramId"`
	InitDataRaw           string   `json:"initDataRaw"`
	SessionID             string   `json:"sessionId"`
	Category              string   `json:"category"`
}

// ThreeDResponse is returned by POST /v1/generate/3d.
type ThreeDResponse struct {
	ModelURL string `json:"model_url"`
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
		TextModels:   ListTextModels(),
		ImageModels:  ListImageModels(),
		VideoModels:  ListVideoModels(),
		AudioModels:  ListAudioModels(),
		ThreeDModels: ListThreeDModels(),
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

	if isKlingVideoModel(def.ID) {
		resolvedID := resolveKlingModelID(req.Resolution, def.ID)
		if resolvedDef, resolvedOK := resolveWavespeedVideoModel(resolvedID); resolvedOK {
			def = resolvedDef
		}
	}

	if err := validateVideoRequest(prompt, req, def); err != nil {
		return VideoResponse{}, err
	}

	return generateWavespeedVideo(ctx, s.cfg, prompt, req, def)
}

func (s *Service) GenerateAudio(ctx context.Context, req AudioRequest) (AudioResponse, error) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(req.Text)
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

	if err := validateAudioRequest(prompt, req, def); err != nil {
		return AudioResponse{}, err
	}

	if isUnifiedQwen3TTS(def) && hasQwen3TTSSourceAudio(req) && strings.TrimSpace(req.SourceAudioURL) == "" && strings.TrimSpace(req.AudioURL) == "" && strings.TrimSpace(req.AudioBase64) == "" {
		return AudioResponse{}, &ProviderError{Provider: "wavespeed", Message: "reference audio is required for voice clone"}
	}

	return generateWavespeedAudio(ctx, s.cfg, prompt, req, def)
}

func validateVideoRequest(prompt string, req VideoRequest, def mediaModelDef) error {
	if def.RequiresImage {
		sourceImage := strings.TrimSpace(req.SourceImageURL)
		if sourceImage == "" {
			sourceImage = strings.TrimSpace(req.ImageURL)
		}
		if sourceImage == "" && strings.TrimSpace(req.ImageBase64) == "" {
			return &ProviderError{Provider: "wavespeed", Message: "source image is required for this model"}
		}
		return nil
	}
	if prompt == "" {
		return ErrPromptEmpty
	}
	return nil
}

func validateAudioRequest(prompt string, req AudioRequest, def mediaModelDef) error {
	switch def.ID {
	case "ace-step-1.5":
		tags := strings.TrimSpace(req.Tags)
		if tags == "" {
			tags = strings.TrimSpace(req.StyleInstruction)
		}
		if tags == "" {
			return ErrPromptEmpty
		}
		return nil
	case "mureka-v9":
		if prompt == "" {
			return ErrPromptEmpty
		}
		return nil
	default:
		if prompt == "" {
			return ErrPromptEmpty
		}
		return nil
	}
}

func (s *Service) GenerateThreeD(ctx context.Context, req ThreeDRequest) (ThreeDResponse, error) {
	prompt := strings.TrimSpace(req.Prompt)
	if prompt == "" {
		prompt = strings.TrimSpace(req.Text)
	}

	if !s.cfg.WavespeedImageEnabled() {
		return ThreeDResponse{}, &ProviderError{
			Provider: "wavespeed",
			Message:  "WAVESPEED_API_KEY is not configured",
		}
	}

	def, ok := resolveWavespeedThreeDModel(req.Model)
	if !ok {
		def, ok = resolveWavespeedThreeDModel("hunyuan3d-v3.1-rapid")
	}
	if !ok {
		return ThreeDResponse{}, &ProviderError{Provider: "wavespeed", Message: "unknown 3d model: " + req.Model}
	}

	if err := validateThreeDRequest(prompt, req, def); err != nil {
		return ThreeDResponse{}, err
	}

	return generateWavespeedThreeD(ctx, s.cfg, prompt, req, def)
}

func validateThreeDRequest(prompt string, req ThreeDRequest, def mediaModelDef) error {
	if def.RequiresMultiImage {
		if len(req.SourceImages) < 2 && len(req.Images) < 2 && len(req.ImageBase64List) < 2 {
			return &ProviderError{Provider: "wavespeed", Message: "at least 2 source images are required"}
		}
		return nil
	}
	if def.RequiresImage {
		if strings.TrimSpace(req.SourceImageURL) == "" &&
			strings.TrimSpace(req.ImageURL) == "" &&
			strings.TrimSpace(req.ImageBase64) == "" {
			return &ProviderError{Provider: "wavespeed", Message: "source image is required"}
		}
		return nil
	}
	if prompt == "" {
		return ErrPromptEmpty
	}
	return nil
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
