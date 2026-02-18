package explore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/Masterminds/semver/v3"
)

type extension struct {
	Module      string   `json:"module"`
	Tier        string   `json:"tier,omitempty"`
	Description string   `json:"description,omitempty"`
	Latest      string   `json:"latest,omitempty"`
	Versions    []string `json:"versions,omitempty"`
	Imports     []string `json:"imports,omitempty"`
	Outputs     []string `json:"outputs,omitempty"`
	Subcommands []string `json:"subcommands,omitempty"`
}

const (
	httpRequestTimeout = 10 * time.Second

	defaultExtensionCatalogURL = "https://registry.k6.io/catalog.json"
)

var errFetchExtensionCatalog = errors.New("failed to fetch extension catalog")

func getDefaultExtensionCatalog(ctx context.Context) (map[string]*extension, error) {
	return getExtensionCatalog(ctx, defaultExtensionCatalogURL)
}

func getExtensionCatalog(ctx context.Context, url string) (map[string]*extension, error) {
	client := &http.Client{Timeout: httpRequestTimeout}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("User-Agent", "xk6-subcommand-explore")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("%w: %s", errFetchExtensionCatalog, resp.Status)
	}

	var catalog map[string]*extension

	err = json.NewDecoder(resp.Body).Decode(&catalog)
	if err != nil {
		return nil, err
	}

	// Update the Latest field for each extension
	for _, ext := range catalog {
		ext.Latest = findLatest(ext.Versions)
	}

	return catalog, nil
}

func findLatest(versions []string) string {
	if len(versions) == 0 {
		return ""
	}

	latest, err := semver.NewVersion(versions[0])
	if err != nil {
		return ""
	}

	for _, v := range versions[1:] {
		ver, err := semver.NewVersion(v)
		if err != nil {
			continue
		}

		if ver.GreaterThan(latest) {
			latest = ver
		}
	}

	return latest.Original()
}
