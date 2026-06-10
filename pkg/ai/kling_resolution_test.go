package ai

import "testing"

func TestResolveKlingModelID(t *testing.T) {
	cases := []struct {
		resolution string
		fallback   string
		want       string
	}{
		{"720p", "kling-v3-pro", "kling-v3-std"},
		{"1080p", "kling-v3-std", "kling-v3-pro"},
		{"4k", "kling-v3-std", "kling-v3-4k"},
		{"", "kling-v3-pro", "kling-v3-pro"},
		{"", "seedance-v1.5-t2v-fast", "kling-v3-std"},
	}

	for _, tc := range cases {
		got := resolveKlingModelID(tc.resolution, tc.fallback)
		if got != tc.want {
			t.Fatalf("resolveKlingModelID(%q, %q) = %q, want %q", tc.resolution, tc.fallback, got, tc.want)
		}
	}
}
