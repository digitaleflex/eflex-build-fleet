package buildrunner

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"time"
)

type Result struct {
	ExitCode int
	Stdout   string
	Stderr   string
	Duration time.Duration
}

type Runner struct {
	WorkRoot string
	Timeout  time.Duration
}

func New(workRoot string, timeout time.Duration) *Runner {
	if workRoot == "" {
		workRoot = os.TempDir()
	}
	if timeout <= 0 {
		timeout = 10 * time.Minute
	}
	return &Runner{WorkRoot: workRoot, Timeout: timeout}
}

// Run executes command in an isolated workspace directory. The workspace is
// created under WorkRoot and removed after execution.
func (r *Runner) Run(ctx context.Context, name string, args ...string) (Result, error) {
	start := time.Now()
	ws, err := os.MkdirTemp(r.WorkRoot, "build-*")
	if err != nil {
		return Result{}, err
	}
	defer os.RemoveAll(ws)

	runCtx, cancel := context.WithTimeout(ctx, r.Timeout)
	defer cancel()

	cmd := exec.CommandContext(runCtx, name, args...)
	cmd.Dir = ws
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	runErr := cmd.Run()
	res := Result{Stdout: stdout.String(), Stderr: stderr.String(), Duration: time.Since(start)}
	if runErr != nil {
		if exitErr, ok := runErr.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
		} else {
			return res, fmt.Errorf("run %s: %w", name, runErr)
		}
	}
	return res, nil
}