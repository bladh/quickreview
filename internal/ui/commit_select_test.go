package ui

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/test"

	"github.com/bladh/quickreview/internal/git"
)

func TestNewCommitClearsChangedFileSelection(t *testing.T) {
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
	if err := os.WriteFile(filepath.Join(dir, "a.txt"), []byte("one\ntwo\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "b.txt"), []byte("new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("git", "add", ".")
	run("git", "commit", "-m", "second")

	repo, err := git.Open(dir)
	if err != nil {
		t.Fatal(err)
	}
	commits, err := repo.Log("HEAD", 10)
	if err != nil || len(commits) != 2 {
		t.Fatalf("log: %v %v", commits, err)
	}

	fyneApp := test.NewApp()
	defer fyneApp.Quit()
	fyneApp.Settings().SetTheme(&ReviewTheme{Light: true})
	a := &App{
		fyneApp:        fyneApp,
		repo:           repo,
		commits:        commits,
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

	a.commitList.OnSelected(0)
	if a.selectedFile != -1 {
		t.Fatalf("selecting a commit selected file %d", a.selectedFile)
	}
	if a.detailTabs.SelectedIndex() != 0 {
		t.Fatalf("commit message tab not selected")
	}
	if len(a.files) < 2 {
		t.Fatalf("expected two changed files, got %d", len(a.files))
	}

	a.fileList.OnSelected(1)
	if a.selectedFile != 1 {
		t.Fatalf("file selection: got %d", a.selectedFile)
	}
	if a.detailTabs.SelectedIndex() != 1 {
		t.Fatalf("diff tab not selected")
	}

	a.commitList.OnSelected(1)
	if a.selectedFile != -1 {
		t.Fatalf("previous file stayed selected: %d", a.selectedFile)
	}
	if a.detailTabs.SelectedIndex() != 0 {
		t.Fatalf("commit message tab not restored")
	}
}
