package explore

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const (
	githubRawHostDefault = "https://raw.githubusercontent.com"
	githubHost           = "github.com"
)

var (
	errUnsupportedRepoHost = errors.New("README fetch only supports github.com repository URLs")
	errInvalidRepoURL      = errors.New("invalid repository URL")
	errReadmeNotFound      = errors.New("README not found in repository")
)

// readmeCandidates is the list of filenames tried at the repository root, in
// order. Most projects use README.md; a few use lowercase or no extension.
//
//nolint:gochecknoglobals
var readmeCandidates = []string{"README.md", "readme.md", "README", "README.MD", "Readme.md"}

// githubRepo holds an owner/name pair extracted from a GitHub repository URL.
type githubRepo struct {
	owner string
	name  string
}

// parseGitHubRepo extracts owner/name from a GitHub repository URL.
//
// Accepts forms like:
//
//	https://github.com/grafana/xk6-faker
//	https://github.com/grafana/xk6-faker/
//	https://github.com/grafana/xk6-faker.git
//	http://github.com/grafana/xk6-faker
//
// Returns errUnsupportedRepoHost for non-github.com hosts.
func parseGitHubRepo(repoURL string) (githubRepo, error) {
	u, err := url.Parse(repoURL)
	if err != nil {
		return githubRepo{}, fmt.Errorf("%w: %w", errInvalidRepoURL, err)
	}

	if u.Host != githubHost && u.Host != "www.github.com" {
		return githubRepo{}, fmt.Errorf("%w: %s", errUnsupportedRepoHost, u.Host)
	}

	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return githubRepo{}, fmt.Errorf("%w: missing owner/name in %s", errInvalidRepoURL, repoURL)
	}

	owner := parts[0]
	name := strings.TrimSuffix(parts[1], ".git")

	return githubRepo{owner: owner, name: name}, nil
}

// fetchReadme downloads the README from a GitHub repository.
//
// It tries each candidate filename in readmeCandidates against the repository's
// HEAD branch, returning the first 200 OK response body. Network failures abort
// immediately; per-file 404s are treated as "try the next candidate".
//
// rawBaseURL allows tests to point at an httptest server; pass an empty string
// to use the public GitHub raw host.
func fetchReadme(ctx context.Context, client *http.Client, rawBaseURL string, repo githubRepo) (string, error) {
	if rawBaseURL == "" {
		rawBaseURL = githubRawHostDefault
	}

	if client == nil {
		client = &http.Client{Timeout: httpRequestTimeout}
	}

	var lastNetErr error

	for _, candidate := range readmeCandidates {
		readmeURL := fmt.Sprintf("%s/%s/%s/HEAD/%s", rawBaseURL, repo.owner, repo.name, candidate)

		body, status, err := httpGetString(ctx, client, readmeURL)
		if err != nil {
			lastNetErr = err

			continue
		}

		if status == http.StatusOK {
			return body, nil
		}
		// non-200 → try next candidate
	}

	if lastNetErr != nil {
		return "", lastNetErr
	}

	return "", errReadmeNotFound
}

// httpGetString performs an authenticated-free GET and returns the body string,
// status code, and a network error if any. A non-2xx status is NOT an error.
func httpGetString(ctx context.Context, client *http.Client, url string) (string, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("User-Agent", "xk6-subcommand-explore")
	req.Header.Set("Accept", "text/plain, text/markdown, */*")

	resp, err := client.Do(req)
	if err != nil {
		return "", 0, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", resp.StatusCode, err
	}

	return string(body), resp.StatusCode, nil
}
