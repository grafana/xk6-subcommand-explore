package explore

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"go.k6.io/k6/v2/cmd/state"
)

// errGHNotAvailable is returned when the `gh` CLI cannot be located on PATH.
// It is a sentinel for the caller to fall back to a different rendering path.
var errGHNotAvailable = errors.New("gh CLI not available on PATH")

// renderReadmeViaGH delegates README rendering to the GitHub CLI.
//
// `gh repo view <owner>/<repo>` does its own TTY detection: it produces
// terminal-styled output (with colours, bold, code-block backgrounds) when
// stdout is a TTY and falls back to plain markdown when piped. To preserve
// that behaviour we pass `gh` the real terminal file descriptors via
// gs.Stdout.RawOutFd / gs.Stderr.RawOutFd; passing gs.Stdout directly would
// hand `gh` a Go pipe and force the non-TTY codepath.
//
// Returns errGHNotAvailable when the `gh` binary cannot be found, allowing the
// caller to fall through to a raw-markdown fetch. Any other non-nil error is
// from `gh` itself (auth issue, repo not found, network error, etc).
func renderReadmeViaGH(ctx context.Context, gs *state.GlobalState, repo githubRepo) error {
	ghPath, err := exec.LookPath("gh")
	if err != nil {
		return errGHNotAvailable
	}

	target := fmt.Sprintf("%s/%s", repo.owner, repo.name)

	// ghPath comes from exec.LookPath; target is composed from a validated GitHub repo.
	cmd := exec.CommandContext(ctx, ghPath, "repo", "view", target) //nolint:gosec

	cmd.Stdout = os.NewFile(uintptr(gs.Stdout.RawOutFd), "stdout") //nolint:gosec
	cmd.Stderr = os.NewFile(uintptr(gs.Stderr.RawOutFd), "stderr") //nolint:gosec
	cmd.Stdin = os.Stdin

	// Forward k6's --no-color preference to gh via the standard NO_COLOR env.
	cmd.Env = os.Environ()
	if gs.Flags.NoColor {
		cmd.Env = append(cmd.Env, "NO_COLOR=1")
	}

	return cmd.Run()
}
