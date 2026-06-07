package ai

import "testing"

func TestSelectSeedanceVideoEditSlug(t *testing.T) {
	turbo := true
	standard := false

	cases := []struct {
		name string
		req  VideoRequest
		want string
	}{
		{
			name: "480p uses standard edit",
			req:  VideoRequest{Resolution: "480p"},
			want: seedanceVideoEditSlug,
		},
		{
			name: "720p uses turbo",
			req:  VideoRequest{Resolution: "720p"},
			want: seedanceVideoEditTurboSlug,
		},
		{
			name: "1080p uses turbo",
			req:  VideoRequest{Resolution: "1080p"},
			want: seedanceVideoEditTurboSlug,
		},
		{
			name: "empty resolution uses standard",
			req:  VideoRequest{},
			want: seedanceVideoEditSlug,
		},
		{
			name: "turbo flag forces turbo",
			req:  VideoRequest{Resolution: "480p", TurboMode: &turbo},
			want: seedanceVideoEditTurboSlug,
		},
		{
			name: "turbo flag off forces standard",
			req:  VideoRequest{Resolution: "1080p", TurboMode: &standard},
			want: seedanceVideoEditSlug,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := selectSeedanceVideoEditSlug(tc.req)
			if got != tc.want {
				t.Fatalf("selectSeedanceVideoEditSlug() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestResolveLegacySeedanceEditTurboAlias(t *testing.T) {
	def, ok := resolveWavespeedVideoModel("seedance-v2-video-edit-turbo")
	if !ok || def.ID != "seedance-v2-video-edit" {
		t.Fatalf("legacy turbo alias: %+v ok=%v", def, ok)
	}
}
