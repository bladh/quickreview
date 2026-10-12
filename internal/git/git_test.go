package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestRepoWorkflow(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=Test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=Test",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("%v: %s", err, out)
		}
	}

	run("git", "init", "-b", "main")
	run("git", "config", "user.email", "test@example.com")
	run("git", "config", "user.name", "Test")

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", "a.txt")
	run("git", "commit", "-m", "first")

	run("git", "checkout", "-b", "feature")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", ".")
	run("git", "commit", "-m", "second")

	repo, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	branches, err := repo.Branches()
	if err != nil || len(branches) < 2 {
		t.Fatalf("branches: %v %v", branches, err)
	}
	commits, err := repo.Log("feature", 10)
	if err != nil || len(commits) != 2 {
		t.Fatalf("log: %v %v", commits, err)
	}
	files, err := repo.ChangedFiles(CommitDiffSpec(commits[0].Hash))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 1 {
		t.Fatalf("expected changed files, got %v", files)
	}
	files, err = repo.ChangedFiles(RangeDiffSpec("main", "feature", true))
	if err != nil {
		t.Fatal(err)
	}
	if len(files) < 1 {
		t.Fatalf("expected range files, got %v", files)
	}
	raw, err := repo.FileDiff(RangeDiffSpec("main", "feature", true), "a.txt")
	if err != nil || raw == "" {
		t.Fatalf("file diff: %q %v", raw, err)
	}
}
