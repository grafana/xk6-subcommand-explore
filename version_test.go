package explore

import (
	"runtime/debug"
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_detectK6Major_fromEnv(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		env  map[string]string
		want int
	}{
		{
			name: "release version",
			env:  map[string]string{"K6_PROVISION_HOST_VERSION": "v3.1.0"},
			want: 3,
		},
		{
			name: "pre-release version",
			env:  map[string]string{"K6_PROVISION_HOST_VERSION": "v2.0.0-rc1"},
			want: 2,
		},
		{
			name: "malformed value falls through to default",
			env:  map[string]string{"K6_PROVISION_HOST_VERSION": "not-a-version"},
			want: defaultK6Major,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := detectK6Major(tt.env, func() (*debug.BuildInfo, bool) {
				return nil, false
			})
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_detectK6Major_envOverridesBuildInfo(t *testing.T) {
	t.Parallel()

	env := map[string]string{"K6_PROVISION_HOST_VERSION": "v3.1.0"}
	info := &debug.BuildInfo{Deps: []*debug.Module{
		{Path: "go.k6.io/k6/v2", Version: "v2.0.0"},
	}}

	got := detectK6Major(env, func() (*debug.BuildInfo, bool) {
		return info, true
	})
	require.Equal(t, 3, got)
}

func Test_detectK6Major_fromBuildInfo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		deps []*debug.Module
		want int
	}{
		{
			name: "k6 v2 via /v2 module path",
			deps: []*debug.Module{
				{Path: "github.com/spf13/cobra", Version: "v1.4.0"},
				{Path: "go.k6.io/k6/v2", Version: "v2.0.0-rc1"},
			},
			want: 2,
		},
		{
			name: "k6 v10 via /v10 module path",
			deps: []*debug.Module{{Path: "go.k6.io/k6/v10", Version: "v10.1.0"}},
			want: 10,
		},
		{
			name: "bare go.k6.io/k6 path ignored (v1 unsupported)",
			deps: []*debug.Module{{Path: "go.k6.io/k6", Version: "v1.6.0"}},
			want: defaultK6Major,
		},
		{
			name: "similar but non-k6 path ignored",
			deps: []*debug.Module{{Path: "go.k6.io/k6/validator", Version: "v1.0.0"}},
			want: defaultK6Major,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := detectK6Major(nil, func() (*debug.BuildInfo, bool) {
				return &debug.BuildInfo{Deps: tt.deps}, true
			})
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_detectK6Major_defaultWhenUnknown(t *testing.T) {
	t.Parallel()

	got := detectK6Major(nil, func() (*debug.BuildInfo, bool) {
		return nil, false
	})
	require.Equal(t, defaultK6Major, got)
}

func Test_catalogURLForVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		major int
		want  string
	}{
		{major: 2, want: "https://registry.k6.io/v2/catalog.json"},
		{major: 3, want: "https://registry.k6.io/v3/catalog.json"},
		{major: 10, want: "https://registry.k6.io/v10/catalog.json"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			t.Parallel()

			require.Equal(t, tt.want, catalogURLForVersion(tt.major))
		})
	}
}
