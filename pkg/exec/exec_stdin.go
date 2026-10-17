package execx

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// RunWithStdin executes a command with the given string piped to its stdin.
// Output is sent directly to the terminal (stdout/stderr).
func (e *Exec) RunWithStdin(ctx context.Context, command string, stdinData string, args ...string) error {
	cmd := e.creator.CommandContext(ctx, command, args...)

	cmd.SetStdin(strings.NewReader(stdinData))
	cmd.SetStdout(os.Stdout)
	cmd.SetStderr(os.Stderr)

	if err := cmd.Run(); err != nil {
		if ctx.Err() != nil {
			return fmt.Errorf("command %q canceled: %w", command, ctx.Err())
		}
		return fmt.Errorf("command %q failed: %w", command, err)
	}

	return nil
}
