package explore

import "strings"

// matchExtensionsByQuery returns extensions whose module path or short name
// (the last path segment) contains query (case-insensitive substring match).
//
// An exact match on the full module path or on the short name is preferred
// over substring matches: when an exact match exists, only it is returned.
// This lets users disambiguate by typing the exact name even when it is a
// prefix or substring of other module paths.
//
// An empty query returns all extensions unchanged.
func matchExtensionsByQuery(extensions []*extension, query string) []*extension {
	if query == "" {
		return extensions
	}

	q := strings.ToLower(query)

	var (
		exact  []*extension
		substr []*extension
	)

	for _, ext := range extensions {
		mod := strings.ToLower(ext.Module)

		short := mod
		if i := strings.LastIndex(mod, "/"); i >= 0 {
			short = mod[i+1:]
		}

		switch {
		case mod == q || short == q:
			exact = append(exact, ext)
		case strings.Contains(mod, q):
			substr = append(substr, ext)
		}
	}

	if len(exact) > 0 {
		return exact
	}

	return substr
}
