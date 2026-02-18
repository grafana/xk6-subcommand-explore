package explore

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
	cmdtests "go.k6.io/k6/cmd/tests"
)

func TestExtensionType(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ext  *extension
		want string
	}{
		{
			name: "javascript extension",
			ext:  &extension{Imports: []string{"k6/x/faker"}},
			want: "js",
		},
		{
			name: "output extension",
			ext:  &extension{Outputs: []string{"json"}},
			want: "out",
		},
		{
			name: "subcommand extension",
			ext:  &extension{Subcommands: []string{"dashboard"}},
			want: "sub",
		},
		{
			name: "no type",
			ext:  &extension{},
			want: "",
		},
		{
			name: "multiple imports",
			ext:  &extension{Imports: []string{"k6/x/faker", "k6/x/other"}},
			want: "js",
		},
		{
			name: "javascript takes precedence",
			ext: &extension{
				Imports: []string{"k6/x/faker"},
				Outputs: []string{"json"},
			},
			want: "js",
		},
		{
			name: "output takes precedence over subcommand",
			ext: &extension{
				Outputs:     []string{"json"},
				Subcommands: []string{"dashboard"},
			},
			want: "out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extensionType(tt.ext)
			if got != tt.want {
				t.Errorf("extensionType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExtensionTier(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ext  *extension
		want string
	}{
		{
			name: "official tier",
			ext:  &extension{Tier: "official"},
			want: "off",
		},
		{
			name: "community tier",
			ext:  &extension{Tier: "community"},
			want: "com",
		},
		{
			name: "empty tier defaults to community",
			ext:  &extension{Tier: ""},
			want: "com",
		},
		{
			name: "unknown tier defaults to community",
			ext:  &extension{Tier: "unknown"},
			want: "com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := extensionTier(tt.ext)
			if got != tt.want {
				t.Errorf("extensionTier() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:funlen
func TestOutputJSON(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		extensions []*extension
		wantErr    bool
	}{
		{
			name: "single extension",
			extensions: []*extension{
				{
					Module:      "github.com/grafana/xk6-faker",
					Tier:        "official",
					Description: "Generate fake data",
					Latest:      "v0.4.4",
					Versions:    []string{"v0.4.4"},
					Imports:     []string{"k6/x/faker"},
				},
			},
			wantErr: false,
		},
		{
			name:       "empty extensions",
			extensions: []*extension{},
			wantErr:    false,
		},
		{
			name: "multiple extensions",
			extensions: []*extension{
				{
					Module:  "github.com/grafana/xk6-faker",
					Tier:    "official",
					Latest:  "v0.4.4",
					Imports: []string{"k6/x/faker"},
				},
				{
					Module:  "github.com/grafana/xk6-tls",
					Tier:    "community",
					Latest:  "v0.1.0",
					Imports: []string{"k6/x/tls"},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := cmdtests.NewGlobalTestState(t)

			err := outputJSON(ts.GlobalState, tt.extensions)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify JSON is valid
				var result []*extension

				err = json.Unmarshal(ts.Stdout.Bytes(), &result)
				require.NoError(t, err)
				require.Len(t, result, len(tt.extensions))
			}
		})
	}
}

//nolint:funlen
func TestOutputTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		extensions []*extension
		brief      bool
		wantErr    bool
	}{
		{
			name: "normal mode",
			extensions: []*extension{
				{
					Module:      "github.com/grafana/xk6-faker",
					Tier:        "official",
					Description: "Generate fake data",
					Latest:      "v0.4.4",
					Imports:     []string{"k6/x/faker"},
				},
			},
			brief:   false,
			wantErr: false,
		},
		{
			name: "brief mode",
			extensions: []*extension{
				{
					Module:      "github.com/grafana/xk6-faker",
					Tier:        "official",
					Description: "Generate fake data",
					Latest:      "v0.4.4",
					Imports:     []string{"k6/x/faker"},
				},
			},
			brief:   true,
			wantErr: false,
		},
		{
			name:       "empty extensions",
			extensions: []*extension{},
			brief:      false,
			wantErr:    false,
		},
		{
			name: "long description truncation",
			extensions: []*extension{
				{
					Module:      "github.com/grafana/xk6-test",
					Tier:        "official",
					Description: "This is a very long description that should be truncated when displayed in the table output because it exceeds the maximum width", //nolint:lll
					Latest:      "v1.0.0",
					Imports:     []string{"k6/x/test"},
				},
			},
			brief:   false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ts := cmdtests.NewGlobalTestState(t)

			err := outputTable(ts.GlobalState, tt.extensions, tt.brief)
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				// Verify output contains expected content
				output := ts.Stdout.String()
				if len(tt.extensions) > 0 {
					require.NotEmpty(t, output, "outputTable() produced empty output")
				}
			}
		})
	}
}

func TestGetTerminalWidth(t *testing.T) {
	t.Parallel()

	// NewGlobalTestState creates a non-TTY stdout by default
	ts := cmdtests.NewGlobalTestState(t)
	got := getTerminalWidth(ts.GlobalState)

	require.Equal(t, defaultTerminalWidth, got)
}
