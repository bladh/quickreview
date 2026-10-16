package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/bladh/quickreview/internal/git"
)

func TestUncommittedChangesEntry(t *testing.T) {
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
	run("git", "branch", "other")
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "new.txt"), []byte("fresh\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	repo, err := git.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	branches, err := repo.Branches()
	if err != nil {
		t.Fatal(err)
	}

	fyneApp := test.NewApp()
	defer fyneApp.Quit()
	fyneApp.Settings().SetTheme(&ReviewTheme{Light: true})
	a := &App{
		fyneApp:        fyneApp,
		repo:           repo,
		branches:       branches,
		selectedCommit: -1,
		selectedFile:   -1,
		hideBinary:     true,
		viewMode:       "commits",
	}
	a.win = fyneApp.NewWindow("test")
	a.build()
	a.win.SetContent(a.layout())
	a.win.Resize(fyne.NewSize(900, 700))
	a.win.Show()

	selectBranch := func(name string) {
		t.Helper()
		for i, b := range a.branches {
			if b.Name == name {
				a.branchList.OnSelected(i)
				return
			}
		}
		t.Fatalf("branch %s not found", name)
	}

	selectBranch("main")
	if len(a.commits) < 2 || !a.commits[0].Uncommitted {
		t.Fatalf("expected uncommitted row first, got %+v", a.commits)
	}
	if a.commits[0].Subject != uncommittedChangesLabel {
		t.Fatalf("label: %q", a.commits[0].Subject)
	}
	if a.commits[1].Subject != "first" {
		t.Fatalf("real commit shifted: %+v", a.commits[1])
	}

	a.commitList.OnSelected(0)
	if a.spec.Mode != "working" {
		t.Fatalf("spec: %+v", a.spec)
	}
	if a.commitPane.subject.Text != uncommittedChangesLabel {
		t.Fatalf("pane: %q", a.commitPane.subject.Text)
	}
	if a.detailTabs.SelectedIndex() != 0 {
		t.Fatal("commit tab not selected")
	}
	paths := map[string]bool{}
	for _, f := range a.files {
		paths[f.Path] = true
	}
	if !paths["a.txt"] || !paths["new.txt"] {
		t.Fatalf("files: %+v", a.files)
	}

	newIdx := -1
	for i, f := range a.files {
		if f.Path == "new.txt" {
			newIdx = i
		}
	}
	a.fileList.OnSelected(newIdx)
	if !strings.Contains(a.diffView.fd.Raw, "+fresh") {
		t.Fatalf("untracked diff: %q", a.diffView.fd.Raw)
	}

	a.commitList.OnSelected(1)
	if a.spec.Mode != "commit" || a.spec.Commit != a.commits[1].Hash {
		t.Fatalf("commit spec: %+v", a.spec)
	}
	if a.selectedFile != -1 {
		t.Fatalf("file stayed selected: %d", a.selectedFile)
	}

	selectBranch("other")
	if len(a.commits) == 0 || a.commits[0].Uncommitted {
		t.Fatalf("other branch should not list working tree changes: %+v", a.commits)
	}
}
