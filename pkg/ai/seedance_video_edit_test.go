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

// TestWavespeedVideoEditModelIDDoesNotDoubleSuffix guards against a
// regression where every RequiresVideo model that wasn't one of the
// seedance-v2 special cases got "-edit" appended to an ID that already
// ended in "-edit"/"-extend" (e.g. "wan-2.7-edit" -> "wan-2.7-edit-edit").
// That bogus ID matched nothing in the price catalog, so billing silently
// fell back to the generic per-category default instead of the model's
// configured price.
func TestWavespeedVideoEditModelIDDoesNotDoubleSuffix(t *testing.T) {
	cases := []struct {
		id   string
		slug string
		want string
	}{
		{id: "wan-2.7-edit", slug: "alibaba/wan-2.7/video-edit", want: "wan-2.7-edit"},
		{id: "happyhorse-video-edit", slug: "alibaba/happyhorse-1.0/video-edit", want: "happyhorse-video-edit"},
		{id: "happyhorse-video-extend", slug: "alibaba/happyhorse-1.0/video-extend", want: "happyhorse-video-extend"},
		{id: "veo-3.1-extend", slug: "google/veo3.1-fast/video-extend", want: "veo-3.1-extend"},
		{id: "seedance-v2-video-extend", slug: "bytedance/seedance-2.0/video-extend", want: "seedance-v2-video-extend"},
		{id: "seedance-v2-video-edit", slug: seedanceVideoEditSlug, want: "seedance-v2-video-edit"},
		{id: "seedance-v2-video-edit", slug: seedanceVideoEditTurboSlug, want: "seedance-v2-video-edit-turbo"},
	}

	for _, tc := range cases {
		def := mediaModelDef{ID: tc.id, RequiresVideo: true}
		got := wavespeedVideoEditModelID(def, tc.slug)
		if got != tc.want {
			t.Fatalf("wavespeedVideoEditModelID(%q, %q) = %q, want %q", tc.id, tc.slug, got, tc.want)
		}
	}
}
