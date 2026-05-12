package explore

import (
	"fmt"
	"regexp"
	"runtime/debug"
	"strconv"
	"strings"
)

const (
	defaultRegistryHost = "https://registry.k6.io"

	// defaultK6Major is the fallback major when no version signal is
	// available. This extension requires k6 v2+ (go.k6.io/k6/v2 in go.mod),
	// so v2 is the only sensible default.
	defaultK6Major = 2
)

// k6ModuleRe matches go.k6.io/k6/vN module paths and captures N. v0 and v1
// of k6 use the bare path "go.k6.io/k6" without a /vN suffix per Go semantic
// import versioning; this extension doesn't load on those hosts, so only
// explicit /vN matches are considered.
var k6ModuleRe = regexp.MustCompile(`^go\.k6\.io/k6/v([1-9][0-9]*)$`)

// detectK6Major returns the active k6 major version. Precedence:
//
//  1. K6_PROVISION_HOST_VERSION env, set by a host k6 binary when it
//     provisions a custom binary that hosts this extension.
//  2. The go.k6.io/k6/vN dependency in build info.
//  3. defaultK6Major.
func detectK6Major(env map[string]string, readBuildInfo func() (*debug.BuildInfo, bool)) int {
	if n := parseMajor(env["K6_PROVISION_HOST_VERSION"]); n > 0 {
		return n
	}

	info, ok := readBuildInfo()
	if !ok {
		return defaultK6Major
	}

	for _, dep := range info.Deps {
		m := k6ModuleRe.FindStringSubmatch(dep.Path)
		if m == nil {
			continue
		}

		n, _ := strconv.Atoi(m[1])

		return n
	}

	return defaultK6Major
}

// catalogURLForVersion returns the registry catalog URL for the given k6 major.
func catalogURLForVersion(major int) string {
	return fmt.Sprintf("%s/v%d/catalog.json", defaultRegistryHost, major)
}

// parseMajor extracts the leading positive integer from "v<N>" or
// "v<N>.<rest>" strings, returning 0 for any other input.
func parseMajor(s string) int {
	s = strings.TrimPrefix(s, "v")
	if i := strings.IndexAny(s, "./-+"); i > 0 {
		s = s[:i]
	}

	n, err := strconv.Atoi(s)
	if err != nil || n <= 0 {
		return 0
	}

	return n
}
