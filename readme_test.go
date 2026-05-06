package explore

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

//nolint:funlen
func TestParseGitHubRepo(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		input     string
		wantOwner string
		wantName  string
		wantErr   error
	}{
		{
			name:      "https url",
			input:     "https://github.com/grafana/xk6-faker",
			wantOwner: "grafana",
			wantName:  "xk6-faker",
		},
		{
			name:      "trailing slash",
			input:     "https://github.com/grafana/xk6-faker/",
			wantOwner: "grafana",
			wantName:  "xk6-faker",
		},
		{
			name:      "git suffix",
			input:     "https://github.com/grafana/xk6-faker.git",
			wantOwner: "grafana",
			wantName:  "xk6-faker",
		},
		{
			name:      "http url upgraded works",
			input:     "http://github.com/grafana/xk6-faker",
			wantOwner: "grafana",
			wantName:  "xk6-faker",
		},
		{
			name:      "www subdomain accepted",
			input:     "https://www.github.com/grafana/xk6-faker",
			wantOwner: "grafana",
			wantName:  "xk6-faker",
		},
		{
			name:    "non-github host rejected",
			input:   "https://gitlab.com/grafana/xk6-faker",
			wantErr: errUnsupportedRepoHost,
		},
		{
			name:    "missing repo name",
			input:   "https://github.com/grafana",
			wantErr: errInvalidRepoURL,
		},
		{
			name:    "empty url",
			input:   "",
			wantErr: errUnsupportedRepoHost,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			repo, err := parseGitHubRepo(tt.input)

			if tt.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tt.wantErr)

				return
			}

			require.NoError(t, err)
			require.Equal(t, tt.wantOwner, repo.owner)
			require.Equal(t, tt.wantName, repo.name)
		})
	}
}

func TestFetchReadmeSuccess(t *testing.T) {
	t.Parallel()

	const want = "# xk6-faker\n\nGenerate fake data."

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only the README.md candidate exists.
		if !strings.HasSuffix(r.URL.Path, "/README.md") {
			http.NotFound(w, r)

			return
		}

		_, _ = w.Write([]byte(want))
	}))
	defer server.Close()

	repo := githubRepo{owner: "grafana", name: "xk6-faker"}

	body, err := fetchReadme(context.Background(), server.Client(), server.URL, repo)

	require.NoError(t, err)
	require.Equal(t, want, body)
}

func TestFetchReadmeFallsBackToLowercase(t *testing.T) {
	t.Parallel()

	const want = "# lower"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// First candidate (README.md) → 404; second (readme.md) → 200.
		if strings.HasSuffix(r.URL.Path, "/README.md") {
			http.NotFound(w, r)

			return
		}

		if strings.HasSuffix(r.URL.Path, "/readme.md") {
			_, _ = w.Write([]byte(want))

			return
		}

		http.NotFound(w, r)
	}))
	defer server.Close()

	body, err := fetchReadme(context.Background(), server.Client(), server.URL, githubRepo{owner: "x", name: "y"})

	require.NoError(t, err)
	require.Equal(t, want, body)
}

func TestFetchReadmeAllCandidatesMissing(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.NotFound(w, nil)
	}))
	defer server.Close()

	_, err := fetchReadme(context.Background(), server.Client(), server.URL, githubRepo{owner: "x", name: "y"})

	require.ErrorIs(t, err, errReadmeNotFound)
}

func TestFetchReadmeContextCancelled(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))
	defer server.Close()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := fetchReadme(ctx, server.Client(), server.URL, githubRepo{owner: "x", name: "y"})

	require.Error(t, err)
}
