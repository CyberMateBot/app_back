package ai

import "testing"

func TestSelectQwen3TTSSlug(t *testing.T) {
	cases := []struct {
		name string
		req  AudioRequest
		want string
	}{
		{
			name: "text to speech",
			req:  AudioRequest{Prompt: "hello"},
			want: qwen3TTSTextSlug,
		},
		{
			name: "voice clone url",
			req:  AudioRequest{Prompt: "hello", SourceAudioURL: "https://cdn.example.com/v.wav"},
			want: qwen3TTSCloneSlug,
		},
		{
			name: "voice clone base64",
			req:  AudioRequest{Prompt: "hello", AudioBase64: "abc123"},
			want: qwen3TTSCloneSlug,
		},
		{
			name: "clone mode",
			req:  AudioRequest{Prompt: "hello", Mode: "clone", SourceAudioURL: "https://x"},
			want: qwen3TTSCloneSlug,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := selectQwen3TTSSlug(tc.req)
			if got != tc.want {
				t.Fatalf("got %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildWavespeedAudioInput_TTS(t *testing.T) {
	def, _ := resolveWavespeedAudioModel("qwen3-tts")
	input := buildWavespeedAudioInput(def, "Привет", AudioRequest{
		Language:         "auto",
		Voice:            "Serena",
		StyleInstruction: "calm",
	}, qwen3TTSTextSlug)

	if input["text"] != "Привет" {
		t.Fatalf("text: %v", input["text"])
	}
	if input["voice"] != "Serena" {
		t.Fatalf("voice: %v", input["voice"])
	}
	if input["style_instruction"] != "calm" {
		t.Fatalf("style: %v", input["style_instruction"])
	}
}

func TestBuildWavespeedAudioInput_Clone(t *testing.T) {
	def, _ := resolveWavespeedAudioModel("qwen3-tts")
	input := buildWavespeedAudioInput(def, "Новый текст", AudioRequest{
		SourceAudioURL: "https://cdn.example.com/ref.wav",
		ReferenceText:  "оригинал",
		Language:       "Russian",
	}, qwen3TTSCloneSlug)

	if input["audio"] != "https://cdn.example.com/ref.wav" {
		t.Fatalf("audio: %v", input["audio"])
	}
	if input["reference_text"] != "оригинал" {
		t.Fatalf("reference_text: %v", input["reference_text"])
	}
}

func TestListAudioModels(t *testing.T) {
	models := ListAudioModels()
	if len(models) < 6 {
		t.Fatalf("expected at least 6 audio models, got %d: %+v", len(models), models)
	}
	found := false
	for _, m := range models {
		if m.ID == "qwen3-tts" {
			found = true
		}
	}
	if !found {
		t.Fatalf("qwen3-tts missing from audio models")
	}
}
