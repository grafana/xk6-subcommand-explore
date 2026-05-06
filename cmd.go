// Package explore contains the xk6-subcommand-explore extension.
package explore

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"go.k6.io/k6/v2/cmd/state"
)

var (
	errMutuallyExclusiveFlags = errors.New("flags --brief, --detailed and --json are mutually exclusive")
	errReadmeNeedsQuery       = errors.New("--readme requires a <query> argument identifying a single extension")
	errReadmeWithJSON         = errors.New("--readme cannot be combined with --json")
	errNoMatch                = errors.New("no extension matched query")
	errAmbiguousQuery         = errors.New("query matched multiple extensions")
)

const (
	helpShort = "Explore k6 extensions for Automatic Resolution"
	helpLong  = `List available k6 extensions from the official extension registry.

Filter extensions by type (javascript, output, subcommand) or tier (official, community).
Supports table output (default) and JSON format for machine-readable output.

Pass a <query> positional argument to focus on a single extension. The query is
matched case-insensitively against the module path; an exact match on the short
name (e.g. xk6-faker) wins over substring matches. Use --readme to also fetch
and display the extension's README from its GitHub repository.

When using the --json flag, the output is an array of extension objects.
Each extension object contains the following properties:

- module (string) The Go module path of the extension
- tier (string) Extension tier: official or community
- description (string) Brief description of the extension's functionality
- latest (string) Latest version tag (e.g., v0.1.0)
- versions (array of strings) All available version tags
- imports (array of strings) JavaScript module import paths (for JavaScript extensions)
- outputs (array of strings) Output type names (for output extensions)
- subcommands (array of strings) Subcommand names (for subcommand extensions)
- repo (object) Repository information including URL

`
	helpExample = `
# List all extensions (table output):
k6 x explore

# Show only module and description columns (brief output):
k6 x explore --brief

# Show full descriptions without truncation:
k6 x explore --no-trunc

# Show detailed information with repository URLs:
k6 x explore --detailed

# Output as JSON (for CI/CD integration):
k6 x explore --json

# Filter by tier or type:
k6 x explore --tier official --type javascript

# Show details for a single extension:
k6 x explore xk6-faker

# Show details + README for a single extension:
k6 x explore xk6-faker --readme
`
)

// newSubcommand creates the "explore" subcommand for the xk6 extension.
func newSubcommand(gs *state.GlobalState) *cobra.Command {
	opts := options{gs: gs}

	cmd := &cobra.Command{
		Use:     "explore [query]",
		Short:   helpShort,
		Long:    helpLong,
		Example: helpExample,
		Args:    cobra.MaximumNArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) == 1 {
				opts.query = args[0]
			}

			return run(opts)
		},
		PreRunE: func(_ *cobra.Command, args []string) error {
			return validateOptions(&opts, args)
		},
	}

	flags := cmd.Flags()

	flags.BoolVar(&opts.json, "json", false, "output in JSON format")
	flags.BoolVar(&opts.brief, "brief", false, "show only module and description columns")
	flags.BoolVar(&opts.detailed, "detailed", false, "output as a list with detailed information")
	flags.BoolVar(&opts.notrunc, "no-trunc", false, "do not truncate descriptions in table output")
	flags.BoolVar(&opts.readme, "readme", false, "fetch and print the extension's README (requires <query>)")
	flags.Var(&opts.tier, "tier", "filter by tier ("+strings.Join(tierValues, ",")+")")
	flags.Var(&opts.kind, "type", "filter by type ("+strings.Join(kindValues, ",")+")")

	return cmd
}

// validateOptions enforces flag combinations that would otherwise be ambiguous.
func validateOptions(opts *options, args []string) error {
	if mutuallyExclusiveOutputs(opts) {
		return errMutuallyExclusiveFlags
	}

	if opts.readme && opts.json {
		return errReadmeWithJSON
	}

	if opts.readme && len(args) == 0 {
		return errReadmeNeedsQuery
	}

	return nil
}

// mutuallyExclusiveOutputs reports whether more than one output-format flag
// (--brief, --detailed, --json) was set.
func mutuallyExclusiveOutputs(opts *options) bool {
	count := 0

	if opts.brief {
		count++
	}

	if opts.detailed {
		count++
	}

	if opts.json {
		count++
	}

	return count > 1
}

func run(opts options) error {
	// use the default catalog URL for now
	// in the future, we could add a flag to specify a custom catalog URL
	catalog, err := getDefaultExtensionCatalog(opts.gs.Ctx)
	if err != nil {
		return err
	}

	extensions := filterExtensions(catalog, opts.kind, opts.tier)

	sortExtensions(extensions)

	if opts.query != "" {
		return runSingle(opts, extensions)
	}

	if opts.json {
		return outputJSON(opts.gs, extensions)
	}

	if opts.detailed {
		return outputDetailed(opts.gs, extensions)
	}

	return outputTable(opts.gs, extensions, opts.brief, opts.notrunc)
}

// runSingle handles the path where the user supplied a <query> argument:
// match against the filtered set, error or disambiguate as needed, then
// render either JSON, detailed view, or detailed view + README.
func runSingle(opts options, extensions []*extension) error {
	matches := matchExtensionsByQuery(extensions, opts.query)

	switch len(matches) {
	case 0:
		return fmt.Errorf("%w: %q", errNoMatch, opts.query)

	case 1:
		// fall through to rendering

	default:
		outputDisambiguation(opts.gs, opts.query, matches)

		return fmt.Errorf("%w: %q matched %d extensions", errAmbiguousQuery, opts.query, len(matches))
	}

	ext := matches[0]

	if opts.json {
		return outputJSON(opts.gs, []*extension{ext})
	}

	var readme string

	if opts.readme {
		body, err := fetchReadmeForExtension(opts.gs.Ctx, ext)
		if err != nil {
			// Don't fail the whole command for a README fetch issue;
			// surface a warning to stderr and continue rendering details.
			_, _ = fmt.Fprintf(opts.gs.Stderr, "warning: could not fetch README: %v\n", err)
		} else {
			readme = body
		}
	}

	return outputSingleWithReadme(opts.gs, ext, readme)
}

// fetchReadmeForExtension extracts the GitHub repo coordinates from the
// extension's catalog entry and downloads its README. Returns an error if the
// repo URL is missing, malformed, non-GitHub, or the README cannot be fetched.
func fetchReadmeForExtension(ctx context.Context, ext *extension) (string, error) {
	if ext.Repo == nil || ext.Repo.URL == "" {
		return "", fmt.Errorf("%w: extension has no repository URL", errInvalidRepoURL)
	}

	repo, err := parseGitHubRepo(ext.Repo.URL)
	if err != nil {
		return "", err
	}

	return fetchReadme(ctx, nil, "", repo)
}

func filterExtensions(catalog map[string]*extension, kind kind, tier tier) []*extension {
	filtered := make([]*extension, 0)

	for _, ext := range catalog {
		if ext.Module == "go.k6.io/k6/v2" {
			continue
		}

		if kind.filter(ext) && tier.filter(ext) {
			filtered = append(filtered, ext)
		}
	}

	return filtered
}

func sortExtensions(extensions []*extension) {
	// Sort filtered extensions by tier (official first),
	// then by type (javascript, output, subcommand),
	// then alphabetically by module name.
	sort.Slice(extensions, func(i, j int) bool {
		// First, sort by tier (official before community)
		if extensions[i].Tier != extensions[j].Tier {
			return extensions[i].Tier > extensions[j].Tier
		}

		// Then, sort by type (javascript, output, subcommand)
		typeI := extensionType(extensions[i])
		typeJ := extensionType(extensions[j])

		if typeI != typeJ {
			return typeI < typeJ
		}

		// Finally, sort alphabetically by module name
		return extensions[i].Module < extensions[j].Module
	})
}
