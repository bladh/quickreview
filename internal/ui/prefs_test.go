package ui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bladh/quickreview/internal/git"
)

func TestBrowseStartPathPrefersDefault(t *testing.T) {
	preferred := t.TempDir()
	fallback := t.TempDir()

	got := browseStartPath(preferred, fallback)
	if got != preferred {
		t.Fatalf("expected preferred %q, got %q", preferred, got)
	}

	missing := filepath.Join(preferred, "does-not-exist-dir")
	got = browseStartPath(missing, fallback)
	// missing path itself fails; parent preferred still exists
	if got != preferred {
		t.Fatalf("expected parent of missing preferred %q, got %q", preferred, got)
	}

	got = browseStartPath("", fallback)
	if got != fallback {
		t.Fatalf("expected fallback %q, got %q", fallback, got)
	}

	got = browseStartPath("/no/such/path/anywhere", "")
	if home, err := os.UserHomeDir(); err == nil && got != home && got == "" {
		t.Fatalf("expected home or empty, got %q", got)
	}
}

func TestDefaultAssetExtensionsExcludeConfig(t *testing.T) {
	set := parseAssetExtensionSet(defaultAssetExtensions)
	for _, ext := range []string{".svg", ".png", ".jpg", ".jpeg", ".wav", ".mp3", ".woff2", ".pdf"} {
		if _, ok := set[ext]; !ok {
			t.Fatalf("expected default asset %q", ext)
		}
	}
	for _, ext := range []string{".xml", ".json", ".yaml", ".yml", ".toml", ".ini", ".conf", ".go", ".md"} {
		if _, ok := set[ext]; ok {
			t.Fatalf("config/source %q should not be a default asset", ext)
		}
	}
}

func TestParseAssetExtensions(t *testing.T) {
	got := parseAssetExtensions("PNG, .Jpg;svg\nwav txt")
	want := []string{".png", ".jpg", ".svg", ".wav", ".txt"}
	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Fatalf("got %v want %v", got, want)
	}
	if !isAssetPath("icons/Logo.PNG", parseAssetExtensionSet("png jpg")) {
		t.Fatal("expected case-insensitive match")
	}
	if isAssetPath("config/app.xml", parseAssetExtensionSet(defaultAssetExtensions)) {
		t.Fatal(".xml must not match default assets")
	}
}

func TestFilterFilesHideAssets(t *testing.T) {
	files := []git.ChangedFile{
		{Path: "main.go"},
		{Path: "icon.png"},
		{Path: "layout.xml"},
		{Path: "blob.bin", Binary: true},
	}
	exts := parseAssetExtensionSet(defaultAssetExtensions)

	got := filterFiles(files, false, true, exts)
	if len(got) != 3 {
		t.Fatalf("hide assets only: got %v", pathsOf(got))
	}
	got = filterFiles(files, true, true, exts)
	if len(got) != 2 || got[0].Path != "main.go" || got[1].Path != "layout.xml" {
		t.Fatalf("hide binary+assets: got %v", pathsOf(got))
	}
	got = filterFiles(files, false, false, exts)
	if len(got) != 4 {
		t.Fatalf("no filters: got %v", pathsOf(got))
	}
}

func pathsOf(files []git.ChangedFile) []string {
	out := make([]string, len(files))
	for i, f := range files {
		out[i] = f.Path
	}
	return out
}
