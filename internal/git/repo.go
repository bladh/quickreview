package git

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Repo wraps a local git repository and shells out to the system git binary.
type Repo struct {
	Path string
}

// Open validates that path is inside a git work tree and returns a Repo rooted
// at the work-tree top level.
func Open(path string) (*Repo, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return nil, err
	}
	info, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		abs = filepath.Dir(abs)
	}

	out, err := run(abs, "rev-parse", "--show-toplevel")
	if err != nil {
		return nil, fmt.Errorf("not a git repository: %w", err)
	}
	return &Repo{Path: strings.TrimSpace(out)}, nil
}

func (r *Repo) run(args ...string) (string, error) {
	return run(r.Path, args...)
}

func run(dir string, args ...string) (string, error) {
	return runAllowing(dir, nil, args...)
}

// runAllowing is like run, but treats the given exit codes as success.
// git diff --no-index exits 1 when the files differ.
func runAllowing(dir string, allow map[int]struct{}, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) && allow != nil {
			if _, ok := allow[exitErr.ExitCode()]; ok {
				return stdout.String(), nil
			}
		}
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return stdout.String(), nil
}

// CurrentBranch returns the checked-out branch name, or HEAD if detached.
func (r *Repo) CurrentBranch() (string, error) {
	out, err := r.run("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// Name returns the directory basename of the repository root.
func (r *Repo) Name() string {
	return filepath.Base(r.Path)
}
