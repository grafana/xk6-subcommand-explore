package explore

import (
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:funlen
func TestMatchExtensionsByQuery(t *testing.T) {
	t.Parallel()

	faker := &extension{Module: "github.com/grafana/xk6-faker"}
	fakerFork := &extension{Module: "github.com/someone/xk6-faker-fork"}
	dashboard := &extension{Module: "github.com/grafana/xk6-dashboard"}
	dns := &extension{Module: "github.com/grafana/xk6-dns"}
	all := []*extension{faker, fakerFork, dashboard, dns}

	tests := []struct {
		name  string
		query string
		want  []*extension
	}{
		{
			name:  "empty query returns all",
			query: "",
			want:  all,
		},
		{
			name:  "exact short name beats substring",
			query: "xk6-faker",
			want:  []*extension{faker},
		},
		{
			name:  "exact match is case-insensitive",
			query: "XK6-Faker",
			want:  []*extension{faker},
		},
		{
			name:  "exact full module path",
			query: "github.com/grafana/xk6-faker",
			want:  []*extension{faker},
		},
		{
			name:  "substring match returns multiple",
			query: "faker",
			want:  []*extension{faker, fakerFork},
		},
		{
			name:  "substring match on owner",
			query: "grafana",
			want:  []*extension{faker, dashboard, dns},
		},
		{
			name:  "no match returns empty",
			query: "nonexistent",
			want:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := matchExtensionsByQuery(all, tt.query)

			require.Len(t, got, len(tt.want))

			for _, want := range tt.want {
				require.Contains(t, got, want)
			}
		})
	}
}
