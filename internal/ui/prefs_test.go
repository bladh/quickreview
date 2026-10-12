package ui

import (
	"os"
	"path/filepath"
	"testing"
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
