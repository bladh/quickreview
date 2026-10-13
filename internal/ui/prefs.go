package ui

import (
	"os"
	"path/filepath"
	"strings"

	"fyne.io/fyne/v2"

	"github.com/bladh/quickreview/internal/git"
)

const (
	prefBrowsePath      = "defaultBrowsePath"
	prefColorScheme     = "colorScheme"
	prefBranchSort      = "branchSort"
	prefAssetExtensions = "assetExtensions"

	SchemeDark  = "dark"
	SchemeLight = "light"

	BranchSortAlpha  = string(git.BranchSortAlpha)
	BranchSortRecent = string(git.BranchSortRecent)

	// defaultAssetExtensions are media/font/binary asset types for "Hide assets".
	// Configuration formats (.xml, .json, .yaml, …) are intentionally omitted.
	defaultAssetExtensions = ".svg .png .jpg .jpeg .gif .webp .ico .bmp .tif .tiff " +
		".wav .mp3 .ogg .flac .aac .m4a .mp4 .webm .mov .avi .mkv " +
		".woff .woff2 .ttf .otf .eot .pdf .psd .ai"
)

// Prefs holds user preferences persisted via Fyne's preference store.
type Prefs struct {
	DefaultBrowsePath string
	ColorScheme       string // SchemeDark or SchemeLight
	BranchSort        string // BranchSortAlpha or BranchSortRecent
	AssetExtensions   string // space/comma-separated extensions, e.g. ".png .jpg"
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
	assets := p.StringWithFallback(prefAssetExtensions, defaultAssetExtensions)
	return Prefs{
		DefaultBrowsePath: browse,
		ColorScheme:       scheme,
		BranchSort:        sortMode,
		AssetExtensions:   normalizeAssetExtensions(assets),
	}
}

func (p Prefs) save(app fyne.App) {
	prefs := app.Preferences()
	prefs.SetString(prefBrowsePath, p.DefaultBrowsePath)
	prefs.SetString(prefColorScheme, p.ColorScheme)
	prefs.SetString(prefBranchSort, p.BranchSort)
	prefs.SetString(prefAssetExtensions, normalizeAssetExtensions(p.AssetExtensions))
}

func (p Prefs) branchSortMode() git.BranchSort {
	if p.BranchSort == BranchSortRecent {
		return git.BranchSortRecent
	}
	return git.BranchSortAlpha
}

// assetExtSet returns normalized extensions (with leading ".") for matching.
func (p Prefs) assetExtSet() map[string]struct{} {
	return parseAssetExtensionSet(p.AssetExtensions)
}

// normalizeAssetExtensions canonicalizes a user-edited extension list.
func normalizeAssetExtensions(s string) string {
	exts := parseAssetExtensions(s)
	if len(exts) == 0 {
		return ""
	}
	return strings.Join(exts, " ")
}

func parseAssetExtensions(s string) []string {
	s = strings.ReplaceAll(s, ",", " ")
	s = strings.ReplaceAll(s, ";", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	parts := strings.Fields(s)
	out := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, p := range parts {
		p = strings.ToLower(strings.TrimSpace(p))
		if p == "" || p == "." {
			continue
		}
		if !strings.HasPrefix(p, ".") {
			p = "." + p
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func parseAssetExtensionSet(s string) map[string]struct{} {
	exts := parseAssetExtensions(s)
	set := make(map[string]struct{}, len(exts))
	for _, e := range exts {
		set[e] = struct{}{}
	}
	return set
}

func isAssetPath(path string, exts map[string]struct{}) bool {
	if len(exts) == 0 || path == "" {
		return false
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == "" {
		return false
	}
	_, ok := exts[ext]
	return ok
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
