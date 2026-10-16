package git

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestUncommittedWorkingTree(t *testing.T) {
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

	repo, err := Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	dirty, err := repo.HasUncommittedChanges()
	if err != nil || dirty {
		t.Fatalf("clean tree: dirty=%v err=%v", dirty, err)
	}
	files, err := repo.ChangedFiles(WorkingDiffSpec())
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 0 {
		t.Fatalf("clean tree files: %+v", files)
	}

	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("fresh\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(dir, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "sub", "c.txt"), []byte("nested"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "staged.txt"), []byte("ready\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", "staged.txt")

	dirty, err = repo.HasUncommittedChanges()
	if err != nil || !dirty {
		t.Fatalf("dirty tree: dirty=%v err=%v", dirty, err)
	}
	files, err = repo.ChangedFiles(WorkingDiffSpec())
	if err != nil {
		t.Fatal(err)
	}
	byPath := map[string]ChangedFile{}
	for _, f := range files {
		byPath[f.Path] = f
	}
	if f, ok := byPath["a.txt"]; !ok || f.Status != StatusModified || f.Additions != 1 {
		t.Fatalf("a.txt: %+v", f)
	}
	if f, ok := byPath["staged.txt"]; !ok || f.Status != StatusAdded || f.Additions != 1 {
		t.Fatalf("staged.txt: %+v", f)
	}
	if f, ok := byPath["new.txt"]; !ok || f.Status != StatusAdded || f.Additions != 1 {
		t.Fatalf("new.txt: %+v", f)
	}
	if f, ok := byPath["sub/c.txt"]; !ok || f.Status != StatusAdded || f.Additions != 1 {
		t.Fatalf("sub/c.txt: %+v", f)
	}

	raw, err := repo.FileDiff(WorkingDiffSpec(), "a.txt")
	if err != nil || !strings.Contains(raw, "+two") {
		t.Fatalf("a.txt diff: %q %v", raw, err)
	}
	raw, err = repo.FileDiff(WorkingDiffSpec(), "new.txt")
	if err != nil || !strings.Contains(raw, "+fresh") {
		t.Fatalf("new.txt diff: %q %v", raw, err)
	}
	raw, err = repo.FileDiff(WorkingDiffSpec(), "sub/c.txt")
	if err != nil || !strings.Contains(raw, "+nested") {
		t.Fatalf("sub/c.txt diff: %q %v", raw, err)
	}
}
