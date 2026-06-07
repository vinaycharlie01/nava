package execx

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"os/exec"

	iox "github.com/nirantaraai/nava/pkg/io"
)

// Commander defines the interface for command execution
type Commander interface {
	CombinedOutput() ([]byte, error)
	Environ() []string
	Output() ([]byte, error)
	Run() error
	Start() error
	StderrPipe() (iox.ReadCloser, error)
	StdinPipe() (iox.WriteCloser, error)
	StdoutPipe() (iox.ReadCloser, error)
	String() string
	Wait() error

	// Field accessors
	SetStdin(stdin iox.Reader)
	SetStdout(stdout iox.Writer)
	SetStderr(stderr iox.Writer)
	SetDir(dir string)
	SetEnv(env []string)
}

// CommandCreator defines the interface for creating commands
type CommandCreator interface {
	CommandContext(ctx context.Context, name string, args ...string) Commander
}

// Executor defines the interface for executing commands
type Executor interface {
	Run(ctx context.Context, command string, streamToLog bool, args ...string) error
	RunInDir(ctx context.Context, dir, command string, streamToLog bool, args ...string) error
}

// ExecCmd wraps *exec.Cmd to implement the Commander interface
type ExecCmd struct {
	*exec.Cmd
}

// CombinedOutput wraps the underlying command's CombinedOutput
func (e *ExecCmd) CombinedOutput() ([]byte, error) {
	return e.Cmd.CombinedOutput()
}

// Environ wraps the underlying command's Environ
func (e *ExecCmd) Environ() []string {
	return e.Cmd.Environ()
}

// Output wraps the underlying command's Output
func (e *ExecCmd) Output() ([]byte, error) {
	return e.Cmd.Output()
}

// Run wraps the underlying command's Run
func (e *ExecCmd) Run() error {
	return e.Cmd.Run()
}

// Start wraps the underlying command's Start
func (e *ExecCmd) Start() error {
	return e.Cmd.Start()
}

// StderrPipe wraps the underlying command's StderrPipe
func (e *ExecCmd) StderrPipe() (iox.ReadCloser, error) {
	return e.Cmd.StderrPipe()
}

// StdinPipe wraps the underlying command's StdinPipe
func (e *ExecCmd) StdinPipe() (iox.WriteCloser, error) {
	return e.Cmd.StdinPipe()
}

// StdoutPipe wraps the underlying command's StdoutPipe
func (e *ExecCmd) StdoutPipe() (iox.ReadCloser, error) {
	return e.Cmd.StdoutPipe()
}

// String wraps the underlying command's String
func (e *ExecCmd) String() string {
	return e.Cmd.String()
}

// Wait wraps the underlying command's Wait
func (e *ExecCmd) Wait() error {
	return e.Cmd.Wait()
}

// SetStdin sets the standard input for the command
func (e *ExecCmd) SetStdin(stdin iox.Reader) {
	e.Cmd.Stdin = stdin
}

// SetStdout sets the standard output for the command
func (e *ExecCmd) SetStdout(stdout iox.Writer) {
	e.Cmd.Stdout = stdout
}

// SetStderr sets the standard error for the command
func (e *ExecCmd) SetStderr(stderr iox.Writer) {
	e.Cmd.Stderr = stderr
}

// SetDir sets the working directory for the command
func (e *ExecCmd) SetDir(dir string) {
	e.Cmd.Dir = dir
}

// SetEnv sets the environment variables for the command
func (e *ExecCmd) SetEnv(env []string) {
	e.Cmd.Env = env
}

// DefaultCommandCreator is the default implementation of CommandCreator
type DefaultCommandCreator struct{}

// CommandContext creates a new ExecCmd
func (d *DefaultCommandCreator) CommandContext(ctx context.Context, name string, args ...string) Commander {
	cmd := exec.CommandContext(ctx, name, args...)
	return &ExecCmd{Cmd: cmd}
}

// Exec is the default implementation of Executor
type Exec struct {
	creator CommandCreator
}

// NewExec creates a new Exec instance with the default command creator
func NewExec() *Exec {
	return &Exec{
		creator: &DefaultCommandCreator{},
	}
}

// NewExecWithCreator creates a new Exec instance with a custom command creator
func NewExecWithCreator(creator CommandCreator) *Exec {
	return &Exec{
		creator: creator,
	}
}

// Run executes a command and streams its output.
// If streamToLog is true, output is sent to slog; otherwise, to terminal.
func (e *Exec) Run(ctx context.Context, command string, streamToLog bool, args ...string) error {
	cmd := e.creator.CommandContext(ctx, command, args...)

	// Set stdin using the interface method
	cmd.SetStdin(os.Stdin)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to get stderr pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start command %q: %w", command, err)
	}

	if streamToLog {
		go streamToSlog(ctx, stdout, slog.LevelInfo)
		go streamToSlog(ctx, stderr, slog.LevelError)
	} else {
		go func() {
			_, _ = io.Copy(os.Stdout, stdout)
		}()
		go func() {
			_, _ = io.Copy(os.Stderr, stderr)
		}()
	}

	// Wait for the command to finish execution
	if err := cmd.Wait(); err != nil {
		// if context was canceled, wrap cleanly
		if ctx.Err() != nil {
			return fmt.Errorf("command %q canceled: %w", command, ctx.Err())
		}
		return fmt.Errorf("command %q failed: %w", command, err)
	}

	return nil
}

// Run is a package-level convenience function that uses the default Exec implementation
func Run(ctx context.Context, command string, streamToLog bool, args ...string) error {
	e := NewExec()
	return e.Run(ctx, command, streamToLog, args...)
}

// streamToSlog reads command output and logs it to slog with the given level.
func streamToSlog(ctx context.Context, r iox.Reader, level slog.Level) {
	scanner := bufio.NewScanner(r)
	const maxCapacity = 1024 * 1024 // 1 MB max line size
	buf := make([]byte, 64*1024)    // 64 KB initial buffer
	scanner.Buffer(buf, maxCapacity)

	for scanner.Scan() {
		select {
		case <-ctx.Done():
			slog.WarnContext(ctx, "stream canceled", "reason", ctx.Err())
			return
		default:
			slog.Log(ctx, level, scanner.Text())
		}
	}
	if err := scanner.Err(); err != nil {
		slog.ErrorContext(ctx, "failed to read stream", "err", err)
	}
}
