package ui

import (
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"

	"github.com/bladh/quickreview/internal/git"
)

const (
	prefBrowsePath  = "defaultBrowsePath"
	prefColorScheme = "colorScheme"
	prefBranchSort  = "branchSort"

	SchemeDark  = "dark"
	SchemeLight = "light"

	BranchSortAlpha  = string(git.BranchSortAlpha)
	BranchSortRecent = string(git.BranchSortRecent)
)

// Prefs holds user preferences persisted via Fyne's preference store.
type Prefs struct {
	DefaultBrowsePath string
	ColorScheme       string // SchemeDark or SchemeLight
	BranchSort        string // BranchSortAlpha or BranchSortRecent
}

func loadPrefs(app fyne.App) Prefs {
	p := app.Preferences()
	scheme := p.StringWithFallback(prefColorScheme, SchemeDark)
	if scheme != SchemeLight {
		scheme = SchemeDark
	}
	browse := p.String(prefBrowsePath)
	if browse == "" {
		if home, err := os.UserHomeDir(); err == nil {
			browse = home
		}
	}
	sortMode := p.StringWithFallback(prefBranchSort, BranchSortAlpha)
	if sortMode != BranchSortRecent {
		sortMode = BranchSortAlpha
	}
	return Prefs{
		DefaultBrowsePath: browse,
		ColorScheme:       scheme,
		BranchSort:        sortMode,
	}
}

func (p Prefs) save(app fyne.App) {
	prefs := app.Preferences()
	prefs.SetString(prefBrowsePath, p.DefaultBrowsePath)
	prefs.SetString(prefColorScheme, p.ColorScheme)
	prefs.SetString(prefBranchSort, p.BranchSort)
}

func (p Prefs) branchSortMode() git.BranchSort {
	if p.BranchSort == BranchSortRecent {
		return git.BranchSortRecent
	}
	return git.BranchSortAlpha
}

func applyTheme(app fyne.App, scheme string) {
	light := scheme == SchemeLight
	applyDiffPalette(light)
	app.Settings().SetTheme(&ReviewTheme{Light: light})
}

// browseStartPath returns the folder Open Repository should start in.
// The Preferences default (preferred) always wins when it exists; fallback is
// only used when no valid preferred path is set.
func browseStartPath(preferred, fallback string) string {
	if dir := existingDir(preferred); dir != "" {
		return dir
	}
	if dir := existingDir(fallback); dir != "" {
		return dir
	}
	if home, err := os.UserHomeDir(); err == nil {
		return home
	}
	return ""
}

func existingDir(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	info, err := os.Stat(path)
	if err == nil && info.IsDir() {
		return path
	}
	dir := filepath.Dir(path)
	if info, err := os.Stat(dir); err == nil && info.IsDir() {
		return dir
	}
	return ""
}
