package explore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestKindSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    kind
		wantErr bool
	}{
		{
			name:    "valid javascript",
			input:   "javascript",
			want:    kindJavaScript,
			wantErr: false,
		},
		{
			name:    "valid output",
			input:   "output",
			want:    kindOutput,
			wantErr: false,
		},
		{
			name:    "valid subcommand",
			input:   "subcommand",
			want:    kindSubcommand,
			wantErr: false,
		},
		{
			name:    "invalid type",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var k kind

			err := k.Set(tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, k)
			}
		})
	}
}

//nolint:nlreturn
func TestKindString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind *kind
		want string
	}{
		{
			name: "javascript",
			kind: func() *kind { k := kindJavaScript; return &k }(),
			want: "javascript",
		},
		{
			name: "output",
			kind: func() *kind { k := kindOutput; return &k }(),
			want: "output",
		},
		{
			name: "subcommand",
			kind: func() *kind { k := kindSubcommand; return &k }(),
			want: "subcommand",
		},
		{
			name: "nil",
			kind: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.kind.String()
			if got != tt.want {
				t.Errorf("kind.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:funlen,nlreturn
func TestKindFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		kind *kind
		ext  *extension
		want bool
	}{
		{
			name: "nil kind matches all",
			kind: nil,
			ext:  &extension{Imports: []string{"k6/x/faker"}},
			want: true,
		},
		{
			name: "javascript matches imports",
			kind: func() *kind { k := kindJavaScript; return &k }(),
			ext:  &extension{Imports: []string{"k6/x/faker"}},
			want: true,
		},
		{
			name: "javascript does not match outputs",
			kind: func() *kind { k := kindJavaScript; return &k }(),
			ext:  &extension{Outputs: []string{"json"}},
			want: false,
		},
		{
			name: "output matches outputs",
			kind: func() *kind { k := kindOutput; return &k }(),
			ext:  &extension{Outputs: []string{"json"}},
			want: true,
		},
		{
			name: "output does not match subcommands",
			kind: func() *kind { k := kindOutput; return &k }(),
			ext:  &extension{Subcommands: []string{"dashboard"}},
			want: false,
		},
		{
			name: "subcommand matches subcommands",
			kind: func() *kind { k := kindSubcommand; return &k }(),
			ext:  &extension{Subcommands: []string{"dashboard"}},
			want: true,
		},
		{
			name: "subcommand does not match imports",
			kind: func() *kind { k := kindSubcommand; return &k }(),
			ext:  &extension{Imports: []string{"k6/x/faker"}},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.kind.filter(tt.ext)
			if got != tt.want {
				t.Errorf("kind.filter() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTierSet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    tier
		wantErr bool
	}{
		{
			name:    "valid official",
			input:   "official",
			want:    tierOfficial,
			wantErr: false,
		},
		{
			name:    "valid community",
			input:   "community",
			want:    tierCommunity,
			wantErr: false,
		},
		{
			name:    "invalid tier",
			input:   "invalid",
			want:    "",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			want:    "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var tr tier

			err := tr.Set(tt.input)

			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				require.Equal(t, tt.want, tr)
			}
		})
	}
}

//nolint:nlreturn
func TestTierString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tier *tier
		want string
	}{
		{
			name: "official",
			tier: func() *tier { t := tierOfficial; return &t }(),
			want: "official",
		},
		{
			name: "community",
			tier: func() *tier { t := tierCommunity; return &t }(),
			want: "community",
		},
		{
			name: "nil",
			tier: nil,
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.tier.String()
			if got != tt.want {
				t.Errorf("tier.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

//nolint:nlreturn
func TestTierFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		tier *tier
		ext  *extension
		want bool
	}{
		{
			name: "nil tier matches all",
			tier: nil,
			ext:  &extension{Tier: "official"},
			want: true,
		},
		{
			name: "official matches official",
			tier: func() *tier { t := tierOfficial; return &t }(),
			ext:  &extension{Tier: "official"},
			want: true,
		},
		{
			name: "official does not match community",
			tier: func() *tier { t := tierOfficial; return &t }(),
			ext:  &extension{Tier: "community"},
			want: false,
		},
		{
			name: "community matches community",
			tier: func() *tier { t := tierCommunity; return &t }(),
			ext:  &extension{Tier: "community"},
			want: true,
		},
		{
			name: "community does not match official",
			tier: func() *tier { t := tierCommunity; return &t }(),
			ext:  &extension{Tier: "official"},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := tt.tier.filter(tt.ext)
			if got != tt.want {
				t.Errorf("tier.filter() = %v, want %v", got, tt.want)
			}
		})
	}
}
