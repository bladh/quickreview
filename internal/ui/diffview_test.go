package ui

import (
	"image/color"
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

func TestUnifiedRowLineNoFollowsPalette(t *testing.T) {
	a := test.NewApp()
	defer a.Quit()

	applyDiffPalette(false)
	row := newUnifiedRowWidget()
	ln := diff.Line{Kind: diff.KindContext, Text: "keep", OldNo: 1, NewNo: 1}
	row.Set(ln)
	if !sameColor(row.lineNoBg.FillColor, color.NRGBA{R: 0x25, G: 0x26, B: 0x2b, A: 0xff}) {
		t.Fatalf("dark lineNoBg: %#v", row.lineNoBg.FillColor)
	}

	applyDiffPalette(true)
	row.Set(ln)
	if !sameColor(row.lineNoBg.FillColor, color.NRGBA{R: 0xf6, G: 0xf8, B: 0xfa, A: 0xff}) {
		t.Fatalf("light lineNoBg stayed dark: %#v", row.lineNoBg.FillColor)
	}
	if !sameColor(row.oldNo.Color, color.NRGBA{R: 0x65, G: 0x6d, B: 0x76, A: 0xff}) {
		t.Fatalf("light lineNoFg: %#v", row.oldNo.Color)
	}
}

func sameColor(got, want color.Color) bool {
	gr, gg, gb, ga := got.RGBA()
	wr, wg, wb, wa := want.RGBA()
	return gr == wr && gg == wg && gb == wb && ga == wa
}
