package explore

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

//nolint:funlen
func TestGetExtensionCatalog(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		catalog    map[string]*extension
		statusCode int
		wantErr    bool
		validate   func(t *testing.T, catalog map[string]*extension)
	}{
		{
			name: "successful fetch with single extension",
			catalog: map[string]*extension{
				"xk6-faker": {
					Module:      "github.com/grafana/xk6-faker",
					Tier:        "official",
					Description: "Generate fake data",
					Versions:    []string{"v0.4.4", "v0.4.3", "v0.4.2"},
					Imports:     []string{"k6/x/faker"},
				},
			},
			statusCode: http.StatusOK,
			wantErr:    false,
			validate: func(t *testing.T, catalog map[string]*extension) {
				t.Helper()
				require.Len(t, catalog, 1)
				ext, ok := catalog["xk6-faker"]
				require.True(t, ok)
				require.Equal(t, "github.com/grafana/xk6-faker", ext.Module)
				require.Equal(t, "v0.4.4", ext.Latest)
			},
		},
		{
			name: "successful fetch with multiple extensions",
			catalog: map[string]*extension{
				"xk6-faker": {
					Module:   "github.com/grafana/xk6-faker",
					Tier:     "official",
					Versions: []string{"v0.4.4"},
					Imports:  []string{"k6/x/faker"},
				},
				"xk6-output-prometheus": {
					Module:   "github.com/grafana/xk6-output-prometheus",
					Tier:     "official",
					Versions: []string{"v1.0.0"},
					Outputs:  []string{"prometheus"},
				},
				"xk6-dashboard": {
					Module:      "github.com/grafana/xk6-dashboard",
					Tier:        "community",
					Versions:    []string{"v0.7.4"},
					Subcommands: []string{"dashboard"},
				},
			},
			statusCode: http.StatusOK,
			wantErr:    false,
			validate: func(t *testing.T, catalog map[string]*extension) {
				t.Helper()
				require.Len(t, catalog, 3)
				require.Equal(t, "v0.4.4", catalog["xk6-faker"].Latest)
				require.Equal(t, "v1.0.0", catalog["xk6-output-prometheus"].Latest)
				require.Equal(t, "v0.7.4", catalog["xk6-dashboard"].Latest)
			},
		},
		{
			name: "versions out of order",
			catalog: map[string]*extension{
				"xk6-test": {
					Module:   "github.com/test/xk6-test",
					Versions: []string{"v0.2.0", "v0.3.0", "v0.1.0"},
				},
			},
			statusCode: http.StatusOK,
			wantErr:    false,
			validate: func(t *testing.T, catalog map[string]*extension) {
				t.Helper()
				require.Equal(t, "v0.3.0", catalog["xk6-test"].Latest)
			},
		},
		{
			name: "extension with no versions",
			catalog: map[string]*extension{
				"xk6-test": {
					Module:   "github.com/test/xk6-test",
					Versions: []string{},
				},
			},
			statusCode: http.StatusOK,
			wantErr:    false,
			validate: func(t *testing.T, catalog map[string]*extension) {
				t.Helper()
				require.Empty(t, catalog["xk6-test"].Latest)
			},
		},
		{
			name:       "http 404 error",
			catalog:    nil,
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
		{
			name:       "http 500 error",
			catalog:    nil,
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name:       "http 403 error",
			catalog:    nil,
			statusCode: http.StatusForbidden,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify User-Agent header
				if r.Header.Get("User-Agent") != "xk6-subcommand-explore" {
					http.Error(w, "Invalid User-Agent", http.StatusInternalServerError)

					return
				}

				w.WriteHeader(tt.statusCode)

				if tt.statusCode == http.StatusOK && tt.catalog != nil {
					err := json.NewEncoder(w).Encode(tt.catalog)
					if err != nil {
						http.Error(w, err.Error(), http.StatusInternalServerError)
					}
				}
			}))
			defer server.Close()

			ctx := context.Background()
			catalog, err := getExtensionCatalog(ctx, server.URL)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.NotNil(t, catalog)

				if tt.validate != nil {
					tt.validate(t, catalog)
				}
			}
		})
	}
}

func TestGetExtensionCatalogInvalidJSON(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	ctx := context.Background()
	catalog, err := getExtensionCatalog(ctx, server.URL)

	require.Error(t, err)
	require.Nil(t, catalog)
}

func TestGetExtensionCatalogContextCancellation(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]*extension{})
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	catalog, err := getExtensionCatalog(ctx, server.URL)

	require.Error(t, err)
	require.Nil(t, catalog)
}

func TestGetExtensionCatalogInvalidURL(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	catalog, err := getExtensionCatalog(ctx, "://invalid-url")

	require.Error(t, err)
	require.Nil(t, catalog)
}

func TestGetExtensionCatalogUnreachableServer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	catalog, err := getExtensionCatalog(ctx, "http://localhost:0")

	require.Error(t, err)
	require.Nil(t, catalog)
}

//nolint:funlen
func TestFindLatest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		versions []string
		want     string
	}{
		{
			name:     "empty versions",
			versions: []string{},
			want:     "",
		},
		{
			name:     "single version",
			versions: []string{"v0.4.4"},
			want:     "v0.4.4",
		},
		{
			name:     "multiple versions in order",
			versions: []string{"v0.4.4", "v0.4.3", "v0.4.2"},
			want:     "v0.4.4",
		},
		{
			name:     "multiple versions out of order",
			versions: []string{"v0.4.2", "v0.4.4", "v0.4.3"},
			want:     "v0.4.4",
		},
		{
			name:     "versions with different major",
			versions: []string{"v0.4.4", "v1.0.0", "v0.5.0"},
			want:     "v1.0.0",
		},
		{
			name:     "versions with pre-release",
			versions: []string{"v0.4.4", "v0.5.0-beta.1", "v0.4.3"},
			want:     "v0.5.0-beta.1",
		},
		{
			name:     "patch versions",
			versions: []string{"v0.4.1", "v0.4.10", "v0.4.2"},
			want:     "v0.4.10",
		},
		{
			name:     "invalid version",
			versions: []string{"invalid"},
			want:     "",
		},
		{
			name:     "first version invalid returns empty",
			versions: []string{"invalid", "v0.4.4", "v0.4.3"},
			want:     "",
		},
		{
			name:     "mix of valid and invalid",
			versions: []string{"v0.4.4", "invalid", "v0.4.3"},
			want:     "v0.4.4",
		},
		{
			name:     "all invalid versions",
			versions: []string{"invalid", "also-invalid"},
			want:     "",
		},
		{
			name:     "version without v prefix",
			versions: []string{"0.4.4", "0.4.3"},
			want:     "0.4.4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := findLatest(tt.versions)
			require.Equal(t, tt.want, got)
		})
	}
}

//nolint:funlen
func TestFilterExtensions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		catalog map[string]*extension
		kind    kind
		tier    tier
		want    int
		check   func(t *testing.T, result []*extension)
	}{
		{
			name: "no filters returns all extensions except k6",
			catalog: map[string]*extension{
				"k6": {
					Module: "go.k6.io/k6",
					Tier:   "official",
				},
				"xk6-faker": {
					Module:  "github.com/grafana/xk6-faker",
					Tier:    "official",
					Imports: []string{"k6/x/faker"},
				},
				"xk6-dashboard": {
					Module:      "github.com/grafana/xk6-dashboard",
					Tier:        "community",
					Subcommands: []string{"dashboard"},
				},
			},
			kind: "",
			tier: "",
			want: 2,
			check: func(t *testing.T, result []*extension) {
				t.Helper()

				// Verify k6 itself is not included
				for _, ext := range result {
					require.NotEqual(t, "go.k6.io/k6", ext.Module)
				}
			},
		},
		{
			name: "filter by javascript type only",
			catalog: map[string]*extension{
				"xk6-faker": {
					Module:  "github.com/grafana/xk6-faker",
					Tier:    "official",
					Imports: []string{"k6/x/faker"},
				},
				"xk6-output-prometheus": {
					Module:  "github.com/grafana/xk6-output-prometheus",
					Tier:    "official",
					Outputs: []string{"prometheus"},
				},
				"xk6-dashboard": {
					Module:      "github.com/grafana/xk6-dashboard",
					Tier:        "community",
					Subcommands: []string{"dashboard"},
				},
			},
			kind: kindJavaScript,
			tier: "",
			want: 1,
			check: func(t *testing.T, result []*extension) {
				t.Helper()
				require.Len(t, result, 1)
				require.NotEmpty(t, result[0].Imports)
			},
		},
		{
			name: "filter by output type only",
			catalog: map[string]*extension{
				"xk6-faker": {
					Module:  "github.com/grafana/xk6-faker",
					Tier:    "official",
					Imports: []string{"k6/x/faker"},
				},
				"xk6-output-prometheus": {
					Module:  "github.com/grafana/xk6-output-prometheus",
					Tier:    "official",
					Outputs: []string{"prometheus"},
				},
			},
			kind: kindOutput,
			tier: "",
			want: 1,
			check: func(t *testing.T, result []*extension) {
				t.Helper()
				require.Len(t, result, 1)
				require.NotEmpty(t, result[0].Outputs)
			},
		},
		{
			name: "filter by subcommand type only",
			catalog: map[string]*extension{
				"xk6-faker": {
					Module:  "github.com/grafana/xk6-faker",
					Tier:    "official",
					Imports: []string{"k6/x/faker"},
				},
				"xk6-dashboard": {
					Module:      "github.com/grafana/xk6-dashboard",
					Tier:        "community",
					Subcommands: []string{"dashboard"},
				},
			},
			kind: kindSubcommand,
			tier: "",
			want: 1,
			check: func(t *testing.T, result []*extension) {
				t.Helper()
				require.Len(t, result, 1)
				require.NotEmpty(t, result[0].Subcommands)
			},
		},
		{
			name: "filter by official tier only",
			catalog: map[string]*extension{
				"xk6-faker": {
					Module:  "github.com/grafana/xk6-faker",
					Tier:    "official",
					Imports: []string{"k6/x/faker"},
				},
				"xk6-dashboard": {
					Module:      "github.com/grafana/xk6-dashboard",
					Tier:        "community",
					Subcommands: []string{"dashboard"},
				},
			},
			kind: "",
			tier: tierOfficial,
			want: 1,
			check: func(t *testing.T, result []*extension) {
				t.Helper()
				require.Len(t, result, 1)
				require.Equal(t, "official", result[0].Tier)
			},
		},
		{
			name: "filter by community tier only",
			catalog: map[string]*extension{
				"xk6-faker": {
					Module:  "github.com/grafana/xk6-faker",
					Tier:    "official",
					Imports: []string{"k6/x/faker"},
				},
				"xk6-dashboard": {
					Module:      "github.com/grafana/xk6-dashboard",
					Tier:        "community",
					Subcommands: []string{"dashboard"},
				},
			},
			kind: "",
			tier: tierCommunity,
			want: 1,
			check: func(t *testing.T, result []*extension) {
				t.Helper()
				require.Len(t, result, 1)
				require.Equal(t, "community", result[0].Tier)
			},
		},
		{
			name: "filter by both kind and tier",
			catalog: map[string]*extension{
				"xk6-faker": {
					Module:  "github.com/grafana/xk6-faker",
					Tier:    "official",
					Imports: []string{"k6/x/faker"},
				},
				"xk6-tls": {
					Module:  "github.com/grafana/xk6-tls",
					Tier:    "community",
					Imports: []string{"k6/x/tls"},
				},
				"xk6-output-prometheus": {
					Module:  "github.com/grafana/xk6-output-prometheus",
					Tier:    "official",
					Outputs: []string{"prometheus"},
				},
			},
			kind: kindJavaScript,
			tier: tierOfficial,
			want: 1,
			check: func(t *testing.T, result []*extension) {
				t.Helper()
				require.Len(t, result, 1)
				require.Equal(t, "official", result[0].Tier)
				require.NotEmpty(t, result[0].Imports)
			},
		},
		{
			name:    "empty catalog returns empty result",
			catalog: map[string]*extension{},
			kind:    "",
			tier:    "",
			want:    0,
			check: func(t *testing.T, result []*extension) {
				t.Helper()
				require.Empty(t, result)
			},
		},
		{
			name: "no matches returns empty result",
			catalog: map[string]*extension{
				"xk6-faker": {
					Module:  "github.com/grafana/xk6-faker",
					Tier:    "official",
					Imports: []string{"k6/x/faker"},
				},
			},
			kind: kindOutput,
			tier: tierCommunity,
			want: 0,
			check: func(t *testing.T, result []*extension) {
				t.Helper()
				require.Empty(t, result)
			},
		},
		{
			name: "k6 module always filtered out",
			catalog: map[string]*extension{
				"k6": {
					Module:  "go.k6.io/k6",
					Tier:    "official",
					Imports: []string{"k6"},
				},
			},
			kind: "",
			tier: "",
			want: 0,
			check: func(t *testing.T, result []*extension) {
				t.Helper()
				require.Empty(t, result)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := filterExtensions(tt.catalog, tt.kind, tt.tier)

			require.Len(t, result, tt.want)

			if tt.check != nil {
				tt.check(t, result)
			}
		})
	}
}
