package explore

import (
	"encoding/json"
	"text/tabwriter"

	"go.k6.io/k6/cmd/state"
	"golang.org/x/term"
)

const (
	normalHeader = "MODULE\tLATEST\tTYPE\tTIER\tDESCRIPTION\n"
	briefHeader  = "MODULE\tDESCRIPTION\n"
	typeColWidth = 4
	tierColWidth = 4
	minDescWidth = 20

	columnPadding = 2

	normalPaddings = 10 // total padding for all columns
	briefPaddings  = 4  // total padding for all columns in brief mode

	defaultTerminalWidth = 120 // default width when not in a terminal

	dots    = "..."
	dotsLen = len(dots)
)

func outputJSON(gs *state.GlobalState, extensions []*extension) error {
	encoder := json.NewEncoder(gs.Stdout)
	encoder.SetIndent("", "  ")

	return encoder.Encode(extensions)
}

func outputTable(gs *state.GlobalState, extensions []*extension, brief bool) error {
	w := tabwriter.NewWriter(gs.Stdout, 0, 0, columnPadding, ' ', 0)
	termWidth := getTerminalWidth(gs)
	otherCols := 0

	// Calculate max description width based on terminal width and other columns
	for _, ext := range extensions {
		otherLen := len(ext.Module)

		if !brief {
			otherLen += len(ext.Latest) + typeColWidth + tierColWidth
		}

		if otherLen > otherCols {
			otherCols = otherLen
		}
	}

	if brief {
		otherCols += briefPaddings
	} else {
		otherCols += normalPaddings
	}

	descWidth := max(termWidth-otherCols, minDescWidth)

	if brief {
		_, _ = w.Write([]byte(briefHeader))
	} else {
		_, _ = w.Write([]byte(normalHeader))
	}

	for _, ext := range extensions {
		module := ext.Module
		latest := ext.Latest
		typ := extensionType(ext)
		tier := extensionTier(ext)

		desc := ext.Description
		if len(desc) > descWidth {
			desc = desc[:descWidth-dotsLen] + dots
		}

		if brief {
			_, _ = w.Write([]byte(module + "\t" + desc + "\n"))

			continue
		}

		_, _ = w.Write([]byte(module + "\t" + latest + "\t" + typ + "\t" + tier + "\t" + desc + "\n"))
	}

	return w.Flush()
}

func extensionType(e *extension) string {
	if len(e.Imports) > 0 {
		return "js"
	}

	if len(e.Outputs) > 0 {
		return "out"
	}

	if len(e.Subcommands) > 0 {
		return "sub"
	}

	return ""
}

func extensionTier(e *extension) string {
	switch e.Tier {
	case "official":
		return "off"
	case "community":
		fallthrough
	default:
		return "com"
	}
}

func getTerminalWidth(gs *state.GlobalState) int {
	if gs.Stdout.IsTTY && term.IsTerminal(gs.Stdout.RawOutFd) {
		width, _, err := term.GetSize(gs.Stdout.RawOutFd)
		if err == nil && width > 0 {
			return width
		}
	}

	return defaultTerminalWidth
}
