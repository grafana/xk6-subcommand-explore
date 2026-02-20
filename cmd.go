// Package explore contains the xk6-subcommand-explore extension.
package explore

import (
	"strings"

	"github.com/spf13/cobra"
	"go.k6.io/k6/cmd/state"
)

// newSubcommand creates the "explore" subcommand for the xk6 extension.
func newSubcommand(gs *state.GlobalState) *cobra.Command {
	opts := options{gs: gs}

	cmd := &cobra.Command{
		Use:   "explore",
		Short: "Explore k6 extensions for Automatic Resolution",
		Long: `List available k6 extensions from the official extension registry.

Filter extensions by type (javascript, output, subcommand) or tier (official, community).
Supports table output (default) and JSON format for machine-readable output.

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
`,
		Example: `
# List all extensions (table output):
k6 x explore

# Show only module and description columns (brief output):
k6 x explore --brief

# Output as JSON (for CI/CD integration):
k6 x explore --json

# Filter by tier or type:
k6 x explore --tier official --type javascript
`,
		RunE: func(_ *cobra.Command, _ []string) error {
			return run(opts)
		},
	}

	flags := cmd.Flags()

	flags.BoolVar(&opts.json, "json", false, "output in JSON format")
	flags.BoolVar(&opts.brief, "brief", false, "show only module and description columns")
	flags.Var(&opts.tier, "tier", "filter by tier ("+strings.Join(tierValues, ",")+")")
	flags.Var(&opts.kind, "type", "filter by type ("+strings.Join(kindValues, ",")+")")

	return cmd
}

func run(opts options) error {
	// use the default catalog URL for now
	// in the future, we could add a flag to specify a custom catalog URL
	catalog, err := getDefaultExtensionCatalog(opts.gs.Ctx)
	if err != nil {
		return err
	}

	extensions := filterExtensions(catalog, opts.kind, opts.tier)

	if opts.json {
		return outputJSON(opts.gs, extensions)
	}

	return outputTable(opts.gs, extensions, opts.brief)
}

func filterExtensions(catalog map[string]*extension, kind kind, tier tier) []*extension {
	filtered := make([]*extension, 0)

	for _, ext := range catalog {
		if ext.Module == "go.k6.io/k6" {
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
