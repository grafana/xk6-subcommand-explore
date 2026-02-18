package explore

import "go.k6.io/k6/subcommand"

func init() {
	subcommand.RegisterExtension("explore", newSubcommand)
}
