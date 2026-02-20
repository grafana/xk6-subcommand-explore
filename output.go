package explore

import (
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/fatih/color"
	"github.com/muesli/reflow/indent"
	"github.com/muesli/reflow/wordwrap"
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

	listMargin = 2
)

func outputJSON(gs *state.GlobalState, extensions []*extension) error {
	encoder := json.NewEncoder(gs.Stdout)
	encoder.SetIndent("", "  ")

	return encoder.Encode(extensions)
}

func outputDetailed(gs *state.GlobalState, extensions []*extension) error {
	heading := color.New(color.Bold).SprintfFunc()
	link := color.New(color.FgBlue, color.Underline).SprintfFunc()
	text := color.New(color.Italic).SprintfFunc()

	if gs.Flags.NoColor {
		heading = fmt.Sprintf
		link = fmt.Sprintf
		text = fmt.Sprintf
	}

	_, _ = fmt.Fprintln(gs.Stdout, heading("Extensions\n----------\n"))

	width := getTerminalWidth(gs) - listMargin

	for _, ext := range extensions {
		module := heading(ext.Module)
		url := link(ext.Repo.URL)
		desc := text(indent.String(wordwrap.String(ext.Description, width), listMargin))

		_, _ = fmt.Fprintf(gs.Stdout, "- %s\n  %s • %s • %s\n  %s\n",
			module, ext.Latest, extensionType(ext), extensionTier(ext), url,
		)
		_, _ = fmt.Fprintln(gs.Stdout, desc)
		_, _ = fmt.Fprintln(gs.Stdout)
	}

	return nil
}

func outputTable(gs *state.GlobalState, extensions []*extension, brief, notrunc bool) error {
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
		typ := abbrev(extensionType(ext))
		tier := abbrev(extensionTier(ext))

		desc := ext.Description
		if !notrunc && len(desc) > descWidth {
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
		return "JavaScript"
	}

	if len(e.Outputs) > 0 {
		return "Output"
	}

	if len(e.Subcommands) > 0 {
		return "Subcommand"
	}

	return ""
}

func extensionTier(e *extension) string {
	switch e.Tier {
	case "official":
		return "Official"
	case "community":
		fallthrough
	default:
		return "Community"
	}
}

func abbrev(s string) string {
	switch s {
	case "JavaScript":
		return "js"
	case "Output":
		return "out"
	case "Subcommand":
		return "sub"
	case "Official":
		return "off"
	case "Community":
		return "com"
	default:
		return s
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
