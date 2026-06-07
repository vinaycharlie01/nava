package execx

import (
	"context"
	"fmt"
	"log/slog"
	"os"
)

// RunInDir executes a command in a specific directory and streams its output.
// If streamToLog is true, output is sent to slog; otherwise, to terminal.
func (e *Exec) RunInDir(ctx context.Context, dir, command string, streamToLog bool, args ...string) error {
	cmd := e.creator.CommandContext(ctx, command, args...)

	// Set the working directory
	cmd.SetDir(dir)

	// Set stdin using the interface method
	cmd.SetStdin(os.Stdin)

	// Force color output for terminal commands
	cmd.SetEnv(append(os.Environ(),
		"FORCE_COLOR=1",
		"COLORTERM=truecolor",
		"TERM=xterm-256color",
	))

	if streamToLog {
		stdout, err := cmd.StdoutPipe()
		if err != nil {
			return fmt.Errorf("failed to get stdout pipe: %w", err)
		}

		stderr, err := cmd.StderrPipe()
		if err != nil {
			return fmt.Errorf("failed to get stderr pipe: %w", err)
		}

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start command %q in %s: %w", command, dir, err)
		}

		go streamToSlog(ctx, stdout, slog.LevelInfo)
		go streamToSlog(ctx, stderr, slog.LevelError)
	} else {
		// Connect directly to terminal for color support
		cmd.SetStdout(os.Stdout)
		cmd.SetStderr(os.Stderr)

		if err := cmd.Start(); err != nil {
			return fmt.Errorf("failed to start command %q in %s: %w", command, dir, err)
		}
	}

	// Wait for the command to finish execution
	if err := cmd.Wait(); err != nil {
		// if context was canceled, wrap cleanly
		if ctx.Err() != nil {
			return fmt.Errorf("command %q in %s canceled: %w", command, dir, ctx.Err())
		}
		return fmt.Errorf("command %q in %s failed: %w", command, dir, err)
	}

	return nil
}
