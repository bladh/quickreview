package ui

import (
	"testing"

	"fyne.io/fyne/v2/test"

	"github.com/bladh/quickreview/internal/diff"
)

func TestDiffViewSplitModeDoesNotPanic(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	raw := `diff --git a/foo.go b/foo.go
--- a/foo.go
+++ b/foo.go
@@ -1,3 +1,4 @@
 package foo
-func Old() {}
+func New() {}
+func Extra() {}
`
	fd := diff.ParseUnified(raw)
	v := NewDiffView()
	w := a.NewWindow("test")
	w.SetContent(v)
	w.Resize(v.MinSize())
	w.Show()

	v.SetDiff(fd)
	v.SetMode(DiffSplit)
	v.SetMode(DiffUnified)
	v.SetMode(DiffSplit)

	// Force layout/render of split row templates.
	_ = newSplitRowWidget()
	row := newSplitRowWidget()
	row.Set(fd.ToSplit()[1])
	row.Resize(row.MinSize())
}
