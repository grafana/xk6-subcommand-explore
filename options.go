package explore

import (
	"errors"

	"go.k6.io/k6/cmd/state"
)

var (
	errInvalidKind = errors.New("invalid type: allowed values are javascript, output, subcommand")
	errInvalidTier = errors.New("invalid tier: allowed values are official, community")
)

type kind string

type tier string

const (
	kindJavaScript kind = "javascript"
	kindOutput     kind = "output"
	kindSubcommand kind = "subcommand"

	tierOfficial  tier = "official"
	tierCommunity tier = "community"
)

//nolint:gochecknoglobals
var (
	kindValues = []string{string(kindJavaScript), string(kindOutput), string(kindSubcommand)}
	tierValues = []string{string(tierOfficial), string(tierCommunity)}
)

func (k *kind) String() string {
	if k == nil {
		return ""
	}

	return string(*k)
}

func (k *kind) Set(s string) error {
	switch kind(s) {
	case kindJavaScript, kindOutput, kindSubcommand:
		*k = kind(s)

		return nil
	default:
		return errInvalidKind
	}
}

func (k *kind) Type() string {
	return "type"
}

func (k *kind) filter(ext *extension) bool {
	if k == nil {
		return true
	}

	var prop []string

	switch *k {
	case kindJavaScript:
		prop = ext.Imports
	case kindOutput:
		prop = ext.Outputs
	case kindSubcommand:
		prop = ext.Subcommands
	default:
		return true
	}

	return len(prop) > 0
}

func (t *tier) String() string {
	if t == nil {
		return ""
	}

	return string(*t)
}

func (t *tier) Set(s string) error {
	switch tier(s) {
	case tierOfficial, tierCommunity:
		*t = tier(s)

		return nil
	default:
		return errInvalidTier
	}
}

func (t *tier) Type() string {
	return "tier"
}

func (t *tier) filter(ext *extension) bool {
	if t == nil {
		return true
	}

	var value bool

	switch *t {
	case tierOfficial:
		value = ext.Tier == "official"
	case tierCommunity:
		value = ext.Tier == "community"
	default:
		return true
	}

	return value
}

type options struct {
	json  bool
	brief bool
	tier  tier
	kind  kind
	gs    *state.GlobalState
}
