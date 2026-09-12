package ai

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestSplitTextIntoChunks_ShortStaysWhole(t *testing.T) {
	got := splitTextIntoChunks("Привет, мир!", 500)
	if len(got) != 1 || got[0] != "Привет, мир!" {
		t.Fatalf("short text should stay a single chunk, got %#v", got)
	}
}

func TestSplitTextIntoChunks_SentenceBoundary(t *testing.T) {
	// Two sentences, ~30 chars each. Limit forces one per chunk.
	text := "Первое предложение здесь. Второе идёт после точки."
	got := splitTextIntoChunks(text, 30)
	if len(got) < 2 {
		t.Fatalf("expected 2+ chunks, got %d: %#v", len(got), got)
	}
	for _, c := range got {
		if utf8.RuneCountInString(c) > 30 {
			t.Errorf("chunk exceeds limit: %q (%d runes)", c, utf8.RuneCountInString(c))
		}
	}
	// The join must still contain the original words in order (whitespace
	// normalization aside).
	joined := strings.Join(got, " ")
	if !strings.Contains(joined, "Первое") || !strings.Contains(joined, "Второе") {
		t.Errorf("join lost content: %q", joined)
	}
}

func TestSplitTextIntoChunks_LongSingleSentence(t *testing.T) {
	// One long sentence without any punctuation — must fall back to hard
	// word wrapping so nothing exceeds the limit.
	text := strings.Repeat("word ", 200) // 5*200 - 1 chars
	limit := 50
	got := splitTextIntoChunks(text, limit)
	if len(got) < 2 {
		t.Fatalf("expected multiple chunks for long sentence, got %d", len(got))
	}
	for _, c := range got {
		if utf8.RuneCountInString(c) > limit {
			t.Errorf("chunk exceeds limit: %q (%d)", c, utf8.RuneCountInString(c))
		}
	}
}

func TestSplitTextIntoChunks_CommaFallback(t *testing.T) {
	// One long sentence with commas — should split at commas when the
	// sentence itself is too big for one chunk.
	text := "aaaa, bbbb, cccc, dddd, eeee, ffff, gggg, hhhh, iiii, jjjj"
	limit := 15
	got := splitTextIntoChunks(text, limit)
	if len(got) < 3 {
		t.Fatalf("expected 3+ chunks via comma split, got %d: %#v", len(got), got)
	}
	for _, c := range got {
		if utf8.RuneCountInString(c) > limit {
			t.Errorf("chunk exceeds limit: %q (%d)", c, utf8.RuneCountInString(c))
		}
	}
}

func TestSplitTextIntoChunks_EmptyReturnsNil(t *testing.T) {
	if got := splitTextIntoChunks("", 100); got != nil {
		t.Fatalf("empty text should produce no chunks, got %#v", got)
	}
	if got := splitTextIntoChunks("hello", 0); got != nil {
		t.Fatalf("zero limit should produce no chunks, got %#v", got)
	}
}

func TestChunkLimitForModel(t *testing.T) {
	cases := []struct {
		def  mediaModelDef
		slug string
		want int
	}{
		{mediaModelDef{ID: "qwen3-tts", TextSlug: qwen3TTSTextSlug}, qwen3TTSTextSlug, 500},
		{mediaModelDef{ID: "qwen3-tts", TextSlug: qwen3TTSTextSlug}, qwen3TTSCloneSlug, 500},
		{mediaModelDef{ID: "omnivoice"}, "any", 500},
		{mediaModelDef{ID: "elevenlabs-v3"}, "any", 1500},
		{mediaModelDef{ID: "minimax-speech-2.6"}, "any", 1500},
		{mediaModelDef{ID: "mureka-v9"}, "any", 0},
	}
	for _, tc := range cases {
		if got := chunkLimitForModel(tc.def, tc.slug); got != tc.want {
			t.Errorf("%s: got %d, want %d", tc.def.ID, got, tc.want)
		}
	}
}
