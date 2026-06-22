package billing

import "testing"

func TestPlanUnlocksModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		plan   string
		model  string
		cat    string
		unlock bool
	}{
		{PlanFree, "yandexgpt", "text", true},
		{PlanFree, "gpt-4o-mini", "text", false},
		{PlanBasic, "gpt-4o-mini", "text", true},
		{PlanBasic, "kling-v3-std", "video", true},
		{PlanBasic, "kling-v3-pro", "video", false},
		{PlanPro, "kling-v3-pro", "video", true},
		{PlanPro, "kling-v3-4k", "video", false},
		{PlanMax, "kling-v3-4k", "video", true},
		{PlanFree, "flux-dev", "image", true},
		{PlanFree, "nano-banana", "image", false},
		{PlanBasic, "nano-banana", "image", true},
		{PlanBasic, "seedream-v4.5", "image", true},
		{PlanBasic, "grok-imagine-edit", "image", false},
		{PlanPro, "grok-imagine-edit", "image", true},
		{PlanBasic, "hailuo-2.3-t2v", "video", true},
		{PlanBasic, "wan-2.7-t2v", "video", false},
		{PlanPro, "wan-2.7-t2v", "video", true},
		{PlanPro, "sora-2-t2v", "video", false},
		{PlanMax, "sora-2-t2v", "video", true},
		{PlanMax, "sora-2-t2v-pro", "video", false},
		{PlanUltra, "sora-2-t2v-pro", "video", true},
		{PlanFree, "kling-v3-std", "video", false},
	}

	for _, tc := range tests {
		got := PlanUnlocksModel(tc.plan, tc.model, tc.cat)
		if got != tc.unlock {
			t.Fatalf("PlanUnlocksModel(%q, %q, %q) = %v, want %v", tc.plan, tc.model, tc.cat, got, tc.unlock)
		}
	}
}

func TestMinPlanForModel(t *testing.T) {
	t.Parallel()

	if got := MinPlanForModel("kling-v3-pro", "video"); got != PlanPro {
		t.Fatalf("got %q, want %q", got, PlanPro)
	}
	if got := MinPlanForModel("unknown-model", "video"); got != PlanBasic {
		t.Fatalf("fallback video got %q, want %q", got, PlanBasic)
	}
}
