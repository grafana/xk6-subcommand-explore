# xk6-subcommand-explore

k6 extension that adds a `k6 x explore` subcommand to browse the official extension registry.

## Architecture

Single-package extension registered via k6's subcommand registration mechanism at init time. The data flow is:

1. Registry fetch: HTTP GET to registry.k6.io/catalog.json, decoded into an in-memory map keyed by extension name.
2. Post-processing: after decoding, the latest version is computed per extension using semver comparison -- the registry does not guarantee version ordering.
3. Filtering and sorting: extensions are filtered by kind/tier flags, then sorted (official before community, then by type, then alphabetically).
4. Output: three mutually exclusive output modes (table, detailed list, JSON) all write to k6's GlobalState stdout, which controls TTY detection and color support.

The extension depends on k6's GlobalState for stdout, stderr, context, and CLI flags (like NoColor). All k6 integration flows through this single dependency.

## Gotchas

- The kind/tier filter types implement both pflag.Value (for CLI binding) and a filter method. A zero-value kind/tier means "no filter" and passes everything through -- but this only works because the filter methods check for empty string, not because Go zero values are inherently safe here. Adding a new filter value without updating both Set() and filter() will silently pass all extensions.
- The catalog fetch hardcodes a User-Agent header that the test server validates. Changing the User-Agent string without updating tests will cause silent test failures that look like HTTP 500 errors, not assertion failures.
- Terminal width detection falls back to a hardcoded default when stdout is not a TTY. Tests that validate table output formatting may produce different column widths in CI versus local runs.
- The sort order uses string comparison on the Tier field where "official" > "community" alphabetically. This is coincidental -- if a third tier is added with a name that sorts differently, the ordering breaks without any compiler warning.
