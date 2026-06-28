package billing

import "testing"

func TestAudioGenerationPrice_MurekaV9(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("mureka-v9", "audio")
	if got := AudioGenerationPrice(base, AudioGenerationParams{ModelID: "mureka-v9", NumberOfSongs: 1}); got != 50 {
		t.Fatalf("1 song: got %d, want 50", got)
	}
	if got := AudioGenerationPrice(base, AudioGenerationParams{ModelID: "mureka-v9", NumberOfSongs: 2}); got != 100 {
		t.Fatalf("2 songs: got %d, want 100", got)
	}
	if got := AudioGenerationPrice(base, AudioGenerationParams{ModelID: "mureka-v9", NumberOfSongs: 3}); got != 150 {
		t.Fatalf("3 songs: got %d, want 150", got)
	}
}

func TestAudioGenerationPrice_ACEStep15(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("ace-step-1.5", "audio")
	if got := AudioGenerationPrice(base, AudioGenerationParams{ModelID: "ace-step-1.5", Duration: 60}); got != 40 {
		t.Fatalf("60s: got %d, want 40", got)
	}
	if got := AudioGenerationPrice(base, AudioGenerationParams{ModelID: "ace-step-1.5", Duration: 30}); got != 20 {
		t.Fatalf("30s: got %d, want 20", got)
	}
	if got := AudioGenerationPrice(base, AudioGenerationParams{ModelID: "ace-step-1.5", Duration: 240}); got != 160 {
		t.Fatalf("240s: got %d, want 160", got)
	}
}

func TestAudioGenerationPrice_Qwen3TTS(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("qwen3-tts", "audio")
	if got := AudioGenerationPrice(base, AudioGenerationParams{ModelID: "qwen3-tts", TextLength: 50}); got != 4 {
		t.Fatalf("50 chars: got %d, want 4", got)
	}
	if got := AudioGenerationPrice(base, AudioGenerationParams{ModelID: "qwen3-tts", TextLength: 500}); got != 20 {
		t.Fatalf("500 chars: got %d, want 20", got)
	}
	if got := AudioGenerationPrice(base, AudioGenerationParams{ModelID: "qwen3-tts", TextLength: 500, VoiceClone: true}); got != 200 {
		t.Fatalf("clone 500 chars: got %d, want 200", got)
	}
}

func TestAudioGenerationPrice_OmniVoice(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("omnivoice", "audio")
	if got := AudioGenerationPrice(base, AudioGenerationParams{ModelID: "omnivoice", TextLength: 500}); got != 15 {
		t.Fatalf("500 chars: got %d, want 15", got)
	}
}

func TestAudioGenerationPrice_ElevenLabsV3(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("elevenlabs-v3", "audio")
	if got := AudioGenerationPrice(base, AudioGenerationParams{ModelID: "elevenlabs-v3", TextLength: 50}); got != 8 {
		t.Fatalf("50 chars (min 1000): got %d, want 8", got)
	}
	if got := AudioGenerationPrice(base, AudioGenerationParams{ModelID: "elevenlabs-v3", TextLength: 2000}); got != 16 {
		t.Fatalf("2000 chars: got %d, want 16", got)
	}
}

func TestAudioGenerationPrice_MiniMaxSpeech(t *testing.T) {
	t.Parallel()

	base := DefaultModelPrice("minimax-speech-2.6", "audio")
	if got := AudioGenerationPrice(base, AudioGenerationParams{ModelID: "minimax-speech-2.6", TextLength: 1000}); got != 3 {
		t.Fatalf("1000 chars: got %d, want 3", got)
	}
	if got := AudioGenerationPrice(base, AudioGenerationParams{ModelID: "minimax-speech-2.6", TextLength: 5000}); got != 15 {
		t.Fatalf("5000 chars: got %d, want 15", got)
	}
}
