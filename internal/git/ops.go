package git

import (
	"fmt"
	"strings"
)

// Fetch runs `git fetch --all --prune`.
func (r *Repo) Fetch() (string, error) {
	return r.run("fetch", "--all", "--prune")
}

// Pull runs `git pull --ff-only` on the current branch.
func (r *Repo) Pull() (string, error) {
	return r.run("pull", "--ff-only")
}

// Push runs `git push` for the current branch.
func (r *Repo) Push() (string, error) {
	return r.run("push")
}

// Checkout switches to an existing local branch (or detached-compatible ref).
func (r *Repo) Checkout(ref string) (string, error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", fmt.Errorf("branch name is required")
	}
	return r.run("checkout", ref)
}

// CheckoutRemote creates/updates a local branch tracking a remote branch.
// remoteBranch should look like "origin/feature".
func (r *Repo) CheckoutRemote(remoteBranch string) (string, error) {
	remoteBranch = strings.TrimSpace(remoteBranch)
	if remoteBranch == "" {
		return "", fmt.Errorf("remote branch is required")
	}
	parts := strings.SplitN(remoteBranch, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("expected remote/branch, got %q", remoteBranch)
	}
	local := parts[1]
	// Prefer modern switch; fall back to checkout -B for older git.
	if out, err := r.run("switch", "-C", local, "--track", remoteBranch); err == nil {
		return out, nil
	}
	return r.run("checkout", "-B", local, "--track", remoteBranch)
}

// CreateBranch creates and checks out a new branch from startPoint (default HEAD).
func (r *Repo) CreateBranch(name, startPoint string) (string, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return "", fmt.Errorf("branch name is required")
	}
	if startPoint == "" {
		startPoint = "HEAD"
	}
	if _, err := r.run("branch", name, startPoint); err != nil {
		return "", err
	}
	return r.run("checkout", name)
}

// Remotes lists configured remote names.
func (r *Repo) Remotes() ([]string, error) {
	out, err := r.run("remote")
	if err != nil {
		return nil, err
	}
	return splitNonEmpty(out), nil
}
