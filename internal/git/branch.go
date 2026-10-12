package git

import (
	"sort"
	"strconv"
	"strings"
	"time"
)

// Branch describes a local or remote branch.
type Branch struct {
	Name       string
	Current    bool
	Remote     bool
	Upstream   string
	CommitTime time.Time
}

// BranchSort controls how Branches results are ordered after load.
type BranchSort string

const (
	BranchSortAlpha  BranchSort = "alpha"
	BranchSortRecent BranchSort = "recent"
)

// Branches lists local and remote branches.
func (r *Repo) Branches() ([]Branch, error) {
	out, err := r.run(
		"for-each-ref",
		"--format=%(refname:short)|%(HEAD)|%(upstream:short)|%(refname)|%(committerdate:unix)",
		"refs/heads",
		"refs/remotes",
	)
	if err != nil {
		return nil, err
	}

	var branches []Branch
	for _, line := range splitNonEmpty(out) {
		parts := strings.SplitN(line, "|", 5)
		if len(parts) < 5 {
			continue
		}
		name := parts[0]
		// Skip remote HEAD symbolic refs like origin/HEAD
		if strings.HasSuffix(name, "/HEAD") {
			continue
		}
		var ts time.Time
		if unix, err := strconv.ParseInt(parts[4], 10, 64); err == nil && unix > 0 {
			ts = time.Unix(unix, 0)
		}
		branches = append(branches, Branch{
			Name:       name,
			Current:    parts[1] == "*",
			Upstream:   parts[2],
			Remote:     strings.HasPrefix(parts[3], "refs/remotes/"),
			CommitTime: ts,
		})
	}
	return branches, nil
}

// LocalBranches returns only local branches.
func LocalBranches(all []Branch) []Branch {
	out := make([]Branch, 0, len(all))
	for _, b := range all {
		if !b.Remote {
			out = append(out, b)
		}
	}
	return out
}

// RemoteBranches returns only remote-tracking branches.
func RemoteBranches(all []Branch) []Branch {
	out := make([]Branch, 0, len(all))
	for _, b := range all {
		if b.Remote {
			out = append(out, b)
		}
	}
	return out
}

// SortBranches orders branches. The current branch is kept first; remaining
// entries follow alphabetically or by newest commit.
func SortBranches(branches []Branch, mode BranchSort) {
	sort.SliceStable(branches, func(i, j int) bool {
		a, b := branches[i], branches[j]
		if a.Current != b.Current {
			return a.Current
		}
		switch mode {
		case BranchSortRecent:
			if !a.CommitTime.Equal(b.CommitTime) {
				return a.CommitTime.After(b.CommitTime)
			}
			return strings.ToLower(a.Name) < strings.ToLower(b.Name)
		default: // alpha
			// Local before remote when names would otherwise interleave oddly.
			if a.Remote != b.Remote {
				return !a.Remote
			}
			return strings.ToLower(a.Name) < strings.ToLower(b.Name)
		}
	})
}

func splitNonEmpty(s string) []string {
	var lines []string
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}
