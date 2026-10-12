package git

import (
	"testing"
	"time"
)

func TestSortBranches(t *testing.T) {
	now := time.Now()
	branches := []Branch{
		{Name: "zeta", CommitTime: now.Add(-2 * time.Hour)},
		{Name: "alpha", Current: true, CommitTime: now.Add(-10 * time.Hour)},
		{Name: "beta", CommitTime: now},
		{Name: "origin/gamma", Remote: true, CommitTime: now.Add(-time.Hour)},
	}

	SortBranches(branches, BranchSortAlpha)
	if !branches[0].Current || branches[0].Name != "alpha" {
		t.Fatalf("current should be first, got %+v", branches[0])
	}
	if branches[1].Name != "beta" || branches[2].Name != "zeta" {
		t.Fatalf("alpha order locals wrong: %+v", branches)
	}

	SortBranches(branches, BranchSortRecent)
	if !branches[0].Current {
		t.Fatal("current should stay first")
	}
	if branches[1].Name != "beta" {
		t.Fatalf("expected newest next, got %+v", branches[1])
	}
}
