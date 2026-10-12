package ui

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/bladh/quickreview/internal/diff"
	"github.com/bladh/quickreview/internal/git"
)

// App is the main review window controller.
type App struct {
	fyneApp fyne.App
	win     fyne.Window
	prefs   Prefs

	repo *git.Repo

	branches []git.Branch
	commits  []git.Commit
	allFiles []git.ChangedFile // full list from git
	files    []git.ChangedFile // visible (possibly filtered)
	spec     git.DiffSpec

	selectedBranch string
	selectedCommit int
	selectedFile   int
	threeDot        bool
	hideBinary      bool
	viewMode        string // "commits" or "compare"
	preferCommitTab bool   // keep Commit tab visible while loading files after commit select

	repoLabel       *widget.Label
	branchList      *widget.List
	commitList      *widget.List
	fileList        *widget.List
	diffView        *DiffView
	commitPane      *CommitPane
	detailTabs      *container.AppTabs
	fileHeader      *widget.Label
	statusLabel     *widget.Label
	reviewTabs      *container.AppTabs
	baseSelect      *widget.Select
	headSelect      *widget.Select
	threeDotCheck   *widget.Check
	hideBinaryCheck *widget.Check
}

// Run creates the window and starts the UI.
func Run(fyneApp fyne.App, initialPath string) {
	prefs := loadPrefs(fyneApp)
	applyTheme(fyneApp, prefs.ColorScheme)
	a := &App{
		fyneApp:        fyneApp,
		prefs:          prefs,
		selectedCommit: -1,
		selectedFile:   -1,
		threeDot:       true,
		hideBinary:     true,
		viewMode:       "commits",
	}
	a.win = fyneApp.NewWindow("QuickReview")
	a.win.Resize(fyne.NewSize(1280, 800))
	a.win.SetMaster()
	a.build()
	a.win.SetMainMenu(a.mainMenu())
	a.win.SetContent(a.layout())
	if initialPath != "" {
		a.openRepo(initialPath)
	}
	a.win.ShowAndRun()
}

func (a *App) mainMenu() *fyne.MainMenu {
	fileMenu := fyne.NewMenu("File",
		fyne.NewMenuItem("Open Repository…", a.pickRepo),
		fyne.NewMenuItem("Refresh", func() {
			if a.repo != nil {
				a.reloadRepo()
			}
		}),
		fyne.NewMenuItem("Working Tree", func() {
			if a.repo == nil {
				return
			}
			a.clearCommitSelection()
			a.commitList.UnselectAll()
			a.spec = git.WorkingDiffSpec()
			a.loadFiles()
			a.showDiffTab()
			a.setStatus("Working tree changes vs HEAD")
		}),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Preferences…", a.showPreferences),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Quit", func() { a.win.Close() }),
	)

	viewMenu := fyne.NewMenu("View",
		fyne.NewMenuItem("Commit Message", a.showCommitTab),
		fyne.NewMenuItem("Diff", a.showDiffTab),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Previous File", func() { a.nudgeFile(-1) }),
		fyne.NewMenuItem("Next File", func() { a.nudgeFile(1) }),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Unified Diff", func() { a.diffView.SetMode(DiffUnified) }),
		fyne.NewMenuItem("Split Diff", func() { a.diffView.SetMode(DiffSplit) }),
	)

	return fyne.NewMainMenu(fileMenu, a.repositoryMenu(), viewMenu)
}

func (a *App) build() {
	a.repoLabel = widget.NewLabel("No repository open")
	a.repoLabel.TextStyle = fyne.TextStyle{Bold: true}

	a.statusLabel = widget.NewLabel("Open a git repository to start reviewing")
	a.statusLabel.Wrapping = fyne.TextTruncate

	a.fileHeader = widget.NewLabel("")
	a.fileHeader.TextStyle = fyne.TextStyle{Monospace: true}
	a.fileHeader.Wrapping = fyne.TextTruncate

	a.diffView = NewDiffView()
	a.commitPane = NewCommitPane(
		func() fyne.Clipboard { return a.win.Clipboard() },
		a.copyStatus,
	)
	a.detailTabs = container.NewAppTabs(
		container.NewTabItem("Commit", a.commitPane),
		container.NewTabItem("Diff", container.NewBorder(a.fileHeader, nil, nil, nil, a.diffView)),
	)

	a.branchList = widget.NewList(
		func() int { return len(a.branches) },
		func() fyne.CanvasObject { return newBranchRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(a.branches) {
				return
			}
			obj.(*branchRow).Set(a.branches[id])
		},
	)
	a.branchList.OnSelected = func(id widget.ListItemID) {
		if id >= len(a.branches) {
			return
		}
		a.selectedBranch = a.branches[id].Name
		a.loadCommits(a.selectedBranch)
	}

	a.commitList = widget.NewList(
		func() int { return len(a.commits) },
		func() fyne.CanvasObject { return newCommitRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(a.commits) {
				return
			}
			obj.(*commitRow).Set(a.commits[id])
		},
	)
	a.commitList.OnSelected = func(id widget.ListItemID) {
		a.selectedCommit = id
		if id < 0 || id >= len(a.commits) {
			return
		}
		a.spec = git.CommitDiffSpec(a.commits[id].Hash)
		a.updateCommitPane()
		a.preferCommitTab = true
		a.loadFiles()
		a.preferCommitTab = false
		a.showCommitTab()
	}

	a.fileList = widget.NewList(
		func() int { return len(a.files) },
		func() fyne.CanvasObject { return newFileRow() },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(a.files) {
				return
			}
			obj.(*fileRow).Set(a.files[id])
		},
	)
	a.fileList.OnSelected = func(id widget.ListItemID) {
		a.selectedFile = id
		a.loadDiff()
		if !a.preferCommitTab {
			a.showDiffTab()
		}
	}

	a.hideBinaryCheck = widget.NewCheck("Hide binary", func(on bool) {
		a.hideBinary = on
		a.applyFileFilter()
	})
	a.hideBinaryCheck.SetChecked(true)

	a.baseSelect = widget.NewSelect(nil, nil)
	a.headSelect = widget.NewSelect(nil, nil)
	a.threeDotCheck = widget.NewCheck("Three-dot (GitHub PR style)", func(on bool) {
		a.threeDot = on
	})
	a.threeDotCheck.SetChecked(true)

	compareBtn := widget.NewButton("Compare", func() {
		a.runCompare()
	})
	compareBtn.Importance = widget.HighImportance

	compareForm := container.NewVBox(
		widget.NewLabelWithStyle("Base", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		a.baseSelect,
		widget.NewLabelWithStyle("Compare", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		a.headSelect,
		a.threeDotCheck,
		widget.NewLabel("Three-dot diffs from the merge base (like GitHub PRs). Two-dot is a direct range."),
		compareBtn,
	)

	a.reviewTabs = container.NewAppTabs(
		container.NewTabItem("Commits", container.NewBorder(
			widget.NewLabel("Commits on selected branch"), nil, nil, nil,
			a.commitList,
		)),
		container.NewTabItem("Compare", container.NewPadded(compareForm)),
	)
	a.reviewTabs.OnSelected = func(tab *container.TabItem) {
		if tab.Text == "Compare" {
			a.viewMode = "compare"
		} else {
			a.viewMode = "commits"
		}
	}
}

func (a *App) layout() fyne.CanvasObject {
	msgBtn := widget.NewButtonWithIcon("Commit", theme.InfoIcon(), a.showCommitTab)
	diffBtn := widget.NewButtonWithIcon("Diff", theme.DocumentIcon(), a.showDiffTab)
	prevBtn := widget.NewButtonWithIcon("", theme.NavigateBackIcon(), func() { a.nudgeFile(-1) })
	nextBtn := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() { a.nudgeFile(1) })
	nav := container.NewHBox(msgBtn, diffBtn, prevBtn, nextBtn)
	top := container.NewBorder(nil, nil, nil, nav, a.repoLabel)

	branchPane := container.NewBorder(
		widget.NewLabelWithStyle("Branches", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		nil, nil, nil, a.branchList,
	)
	filesHeader := container.NewBorder(nil, nil,
		widget.NewLabelWithStyle("Changed files", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		a.hideBinaryCheck, nil)
	filesPane := container.NewBorder(filesHeader, nil, nil, nil, a.fileList)

	// Three movable panes under the detail view: Branches | Commits/Compare | Files
	bottomRight := container.NewHSplit(a.reviewTabs, filesPane)
	bottomRight.SetOffset(0.55)
	bottom := container.NewHSplit(branchPane, bottomRight)
	bottom.SetOffset(0.28)

	// Commit/Diff tabs take full width on top; navigation panes share the space below.
	main := container.NewVSplit(a.detailTabs, bottom)
	main.SetOffset(0.62)

	return container.NewBorder(top, a.statusLabel, nil, nil, main)
}

func (a *App) pickRepo() {
	d := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
		if err != nil {
			dialog.ShowError(err, a.win)
			return
		}
		if uri == nil {
			return
		}
		a.openRepo(uri.Path())
	}, a.win)
	fallback := ""
	if a.repo != nil {
		fallback = filepath.Dir(a.repo.Path)
	}
	start := browseStartPath(a.prefs.DefaultBrowsePath, fallback)
	if start != "" {
		if u, err := storage.ListerForURI(storage.NewFileURI(start)); err == nil {
			d.SetLocation(u)
		}
	}
	// Show must run before Resize — Fyne's FileDialog internals are nil until shown.
	d.Show()
	d.Resize(fyne.NewSize(800, 560))
}

func (a *App) openRepo(path string) {
	repo, err := git.Open(path)
	if err != nil {
		dialog.ShowError(err, a.win)
		return
	}
	a.repo = repo
	a.repoLabel.SetText(fmt.Sprintf("%s  —  %s", repo.Name(), repo.Path))
	a.reloadRepo()
}

func (a *App) reloadRepo() {
	branches, err := a.repo.Branches()
	if err != nil {
		dialog.ShowError(err, a.win)
		return
	}
	git.SortBranches(branches, a.prefs.branchSortMode())
	a.branches = branches
	a.branchList.Refresh()

	names := make([]string, 0, len(branches))
	current := ""
	for _, b := range branches {
		names = append(names, b.Name)
		if b.Current {
			current = b.Name
		}
	}
	a.baseSelect.Options = names
	a.headSelect.Options = names
	a.baseSelect.Refresh()
	a.headSelect.Refresh()
	if current != "" {
		a.headSelect.SetSelected(current)
		a.baseSelect.SetSelected(preferDefaultBase(names, current))
		a.selectedBranch = current
		for i, b := range branches {
			if b.Name == current {
				a.branchList.Select(i)
				break
			}
		}
		a.loadCommits(current)
	}
	a.setStatus("Loaded %d branches", len(branches))
}

func preferDefaultBase(names []string, current string) string {
	for _, cand := range []string{"main", "master", "develop", "trunk"} {
		for _, n := range names {
			if n == cand && n != current {
				return n
			}
			if strings.HasSuffix(n, "/"+cand) && n != current {
				return n
			}
		}
	}
	for _, n := range names {
		if n != current {
			return n
		}
	}
	return current
}

func (a *App) loadCommits(ref string) {
	commits, err := a.repo.Log(ref, 200)
	if err != nil {
		dialog.ShowError(err, a.win)
		return
	}
	a.commits = commits
	a.commitList.UnselectAll()
	a.commitList.Refresh()
	a.clearCommitSelection()
	a.clearFiles()
	a.setStatus("%s — %d commits", ref, len(commits))
}

func (a *App) runCompare() {
	base := a.baseSelect.Selected
	head := a.headSelect.Selected
	if base == "" || head == "" {
		dialog.ShowInformation("Compare", "Select both base and compare refs.", a.win)
		return
	}
	a.spec = git.RangeDiffSpec(base, head, a.threeDot)
	commits, err := a.repo.LogRange(base, head, 200)
	if err != nil {
		dialog.ShowError(err, a.win)
		return
	}
	a.commits = commits
	a.commitList.Refresh()
	a.reviewTabs.SelectIndex(0)
	a.loadFiles()
	sep := "..."
	if !a.threeDot {
		sep = ".."
	}
	a.setStatus("Comparing %s%s%s — %d commits", base, sep, head, len(commits))
}

func (a *App) loadFiles() {
	files, err := a.repo.ChangedFiles(a.spec)
	if err != nil {
		dialog.ShowError(err, a.win)
		return
	}
	a.allFiles = files
	a.applyFileFilter()
}

func (a *App) applyFileFilter() {
	prevPath := ""
	if a.selectedFile >= 0 && a.selectedFile < len(a.files) {
		prevPath = a.files[a.selectedFile].Path
	}

	a.files = filterFiles(a.allFiles, a.hideBinary)
	a.selectedFile = -1
	a.fileList.UnselectAll()
	a.fileList.Refresh()

	hidden := len(a.allFiles) - len(a.files)
	switch {
	case a.hideBinary && hidden > 0:
		a.setStatus("%d changed files (%d binary hidden)", len(a.files), hidden)
	default:
		a.setStatus("%d changed files", len(a.files))
	}

	if len(a.files) == 0 {
		a.diffView.Clear()
		a.fileHeader.SetText("")
		return
	}

	selectIdx := 0
	if prevPath != "" {
		for i, f := range a.files {
			if f.Path == prevPath {
				selectIdx = i
				break
			}
		}
	}
	a.fileList.Select(selectIdx)
}

func filterFiles(files []git.ChangedFile, hideBinary bool) []git.ChangedFile {
	if !hideBinary {
		out := make([]git.ChangedFile, len(files))
		copy(out, files)
		return out
	}
	out := make([]git.ChangedFile, 0, len(files))
	for _, f := range files {
		if !f.Binary {
			out = append(out, f)
		}
	}
	return out
}

func (a *App) clearFiles() {
	a.allFiles = nil
	a.files = nil
	a.selectedFile = -1
	a.fileList.Refresh()
	a.diffView.Clear()
	a.fileHeader.SetText("")
}

func (a *App) clearCommitSelection() {
	a.selectedCommit = -1
	a.updateCommitPane()
}

func (a *App) loadDiff() {
	if a.selectedFile < 0 || a.selectedFile >= len(a.files) {
		return
	}
	f := a.files[a.selectedFile]
	raw, err := a.repo.FileDiff(a.spec, f.Path)
	if err != nil {
		dialog.ShowError(err, a.win)
		return
	}
	fd := diff.ParseUnified(raw)
	if f.Binary {
		fd.Binary = true
	}
	a.diffView.SetDiff(fd)

	title := f.Path
	if f.OldPath != "" && f.OldPath != f.Path {
		title = f.OldPath + " → " + f.Path
	}
	a.fileHeader.SetText(fmt.Sprintf("%s  [%s]  %s",
		title, git.StatusLabel(f.Status), FormatChurn(f.Additions, f.Deletions, f.Binary)))
	a.setStatus("Reviewing %s", filepath.Base(f.Path))
}

func (a *App) nudgeFile(delta int) {
	if len(a.files) == 0 {
		return
	}
	next := a.selectedFile + delta
	if a.selectedFile < 0 {
		next = 0
	}
	if next < 0 {
		next = 0
	}
	if next >= len(a.files) {
		next = len(a.files) - 1
	}
	a.fileList.Select(next)
}

func (a *App) setStatus(format string, args ...any) {
	a.statusLabel.SetText(fmt.Sprintf(format, args...))
}

func formatDate(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Local().Format("2006-01-02 15:04")
}
