package diff

import (
	"testing"
)

func TestParseUnifiedAndSplit(t *testing.T) {
	raw := `diff --git a/foo.go b/foo.go
index 111..222 100644
--- a/foo.go
+++ b/foo.go
@@ -1,4 +1,5 @@
 package foo
-func Old() {}
+func New() {}
+func Extra() {}
 const X = 1
`
	fd := ParseUnified(raw)
	if fd.OldFile != "foo.go" || fd.NewFile != "foo.go" {
		t.Fatalf("paths: old=%q new=%q", fd.OldFile, fd.NewFile)
	}
	if len(fd.Hunks) != 1 {
		t.Fatalf("hunks: %d", len(fd.Hunks))
	}
	flat := fd.FlatLines()
	if len(flat) < 5 {
		t.Fatalf("flat lines: %d", len(flat))
	}
	rows := fd.ToSplit()
	if len(rows) < 4 {
		t.Fatalf("split rows: %d", len(rows))
	}
	// Expect paired delete/add for Old→New
	foundPair := false
	for _, r := range rows {
		if r.OldKind == KindDel && r.NewKind == KindAdd {
			foundPair = true
			break
		}
	}
	if !foundPair {
		t.Fatal("expected a delete/add split pair")
	}
}
