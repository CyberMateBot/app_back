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
