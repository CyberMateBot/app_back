package applinks

import "testing"

func TestReferralLink(t *testing.T) {
	t.Parallel()

	tests := []struct {
		bot      string
		telegram string
		want     string
	}{
		{"CyberMate_bot", "12345", "https://t.me/CyberMate_bot?startapp=ref_12345"},
		{"@CyberMate_bot", "999", "https://t.me/CyberMate_bot?startapp=ref_999"},
		{"", "123", ""},
		{"CyberMate_bot", "", ""},
	}

	for _, tc := range tests {
		if got := ReferralLink(tc.bot, tc.telegram, ""); got != tc.want {
			t.Fatalf("ReferralLink(%q, %q) = %q, want %q", tc.bot, tc.telegram, got, tc.want)
		}
	}
}

func TestParseReferralStartParam(t *testing.T) {
	t.Parallel()

	if got := ParseReferralStartParam("ref_777000", ""); got != "777000" {
		t.Fatalf("got %q", got)
	}
	if got := ParseReferralStartParam("777000", ""); got != "777000" {
		t.Fatalf("got %q", got)
	}
}

func TestMainMiniAppFullscreenURL(t *testing.T) {
	t.Parallel()

	if got := MainMiniAppFullscreenURL("CyberMate_bot"); got != "https://t.me/CyberMate_bot?startapp&mode=fullscreen" {
		t.Fatalf("got %q", got)
	}
	if got := MainMiniAppFullscreenURL(""); got != "" {
		t.Fatalf("got %q", got)
	}
}
