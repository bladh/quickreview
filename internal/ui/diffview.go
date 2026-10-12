package ui

import (
	"fmt"
	"image/color"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/bladh/quickreview/internal/diff"
)

// DiffMode is unified or split side-by-side.
type DiffMode int

const (
	DiffUnified DiffMode = iota
	DiffSplit
)

const rowHeight float32 = 16

// DiffView renders a parsed file diff in unified or split mode.
type DiffView struct {
	widget.BaseWidget
	mode DiffMode
	fd   diff.FileDiff

	unified []diff.Line
	split   []diff.SplitRow

	list   *widget.List
	empty  *canvas.Text
	stack  *fyne.Container
}

func NewDiffView() *DiffView {
	v := &DiffView{mode: DiffUnified}
	v.ExtendBaseWidget(v)

	v.empty = canvas.NewText("Select a changed file to review", colorLineNoFg)
	v.empty.TextSize = 13
	v.empty.Alignment = fyne.TextAlignCenter

	v.list = widget.NewList(
		func() int { return v.rowCount() },
		func() fyne.CanvasObject { return v.createRowTemplate() },
		func(id widget.ListItemID, obj fyne.CanvasObject) { v.updateRow(id, obj) },
	)
	v.list.HideSeparators = true

	v.stack = container.NewStack(v.empty, v.list)
	v.showEmpty("Select a changed file to review")
	return v
}

func (v *DiffView) SetMode(mode DiffMode) {
	if v.mode == mode {
		return
	}
	v.mode = mode
	// Row templates differ by mode; rebuild the list so items match.
	v.list = widget.NewList(
		func() int { return v.rowCount() },
		func() fyne.CanvasObject { return v.createRowTemplate() },
		func(id widget.ListItemID, obj fyne.CanvasObject) { v.updateRow(id, obj) },
	)
	v.list.HideSeparators = true
	v.stack.Objects = []fyne.CanvasObject{v.empty, v.list}
	if v.fd.Binary || len(v.fd.Hunks) > 0 || len(v.unified) > 0 || len(v.split) > 0 {
		v.SetDiff(v.fd)
	} else {
		v.showEmpty("Select a changed file to review")
		v.stack.Refresh()
	}
}

func (v *DiffView) Mode() DiffMode { return v.mode }

func (v *DiffView) SetDiff(fd diff.FileDiff) {
	v.fd = fd
	v.unified = fd.FlatLines()
	v.split = fd.ToSplit()

	switch {
	case fd.Binary:
		v.showEmpty("Binary file — content not shown")
	case v.rowCount() == 0:
		v.showEmpty("No textual changes in this file")
	default:
		v.empty.Hide()
		v.list.Show()
		v.stack.Refresh()
		// MinSize forces CreateRenderer; ScrollToTop panics if scroller is still nil.
		_ = v.list.MinSize()
		v.list.Refresh()
		v.list.ScrollToTop()
	}
}

func (v *DiffView) Clear() {
	v.fd = diff.FileDiff{}
	v.unified = nil
	v.split = nil
	v.showEmpty("Select a changed file to review")
}

func (v *DiffView) showEmpty(msg string) {
	v.empty.Text = msg
	v.empty.Refresh()
	v.empty.Show()
	v.list.Hide()
}

func (v *DiffView) rowCount() int {
	if v.mode == DiffSplit {
		return len(v.split)
	}
	return len(v.unified)
}

func (v *DiffView) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(v.stack)
}

func (v *DiffView) createRowTemplate() fyne.CanvasObject {
	if v.mode == DiffSplit {
		return newSplitRowWidget()
	}
	return newUnifiedRowWidget()
}

func (v *DiffView) updateRow(id widget.ListItemID, obj fyne.CanvasObject) {
	if v.mode == DiffSplit {
		if id >= len(v.split) {
			return
		}
		if row, ok := obj.(*splitRowWidget); ok {
			row.Set(v.split[id])
			return
		}
		// Mode switched — rebuild template content by replacing isn't possible;
		// force list refresh via matching type: fall through recreate.
		return
	}
	if id >= len(v.unified) {
		return
	}
	if row, ok := obj.(*unifiedRowWidget); ok {
		row.Set(v.unified[id])
	}
}

// --- unified row ---

type unifiedRowWidget struct {
	widget.BaseWidget
	bg           *canvas.Rectangle
	oldNo, newNo *canvas.Text
	gutter       *canvas.Text
	code         *canvas.Text
	content      *fyne.Container
}

func newUnifiedRowWidget() *unifiedRowWidget {
	r := &unifiedRowWidget{
		bg:     canvas.NewRectangle(colorCtxBg),
		oldNo:  canvas.NewText("", colorLineNoFg),
		newNo:  canvas.NewText("", colorLineNoFg),
		gutter: canvas.NewText(" ", theme.ForegroundColor()),
		code:   canvas.NewText("", theme.ForegroundColor()),
	}
	r.oldNo.TextSize, r.newNo.TextSize = 11, 11
	r.oldNo.TextStyle = fyne.TextStyle{Monospace: true}
	r.newNo.TextStyle = fyne.TextStyle{Monospace: true}
	r.oldNo.Alignment = fyne.TextAlignTrailing
	r.newNo.Alignment = fyne.TextAlignTrailing
	r.gutter.TextSize = 12
	r.gutter.TextStyle = fyne.TextStyle{Monospace: true}
	r.gutter.Alignment = fyne.TextAlignCenter
	r.code.TextSize = 12
	r.code.TextStyle = fyne.TextStyle{Monospace: true}

	lineNoBg := canvas.NewRectangle(colorLineNoBg)
	nums := container.New(layout.NewGridLayout(2), r.oldNo, r.newNo)
	left := container.NewStack(lineNoBg, nums)
	gutterBg := canvas.NewRectangle(colorCtxBg)
	gutterBox := container.NewStack(gutterBg, container.NewCenter(r.gutter))

	r.content = container.NewStack(r.bg,
		container.NewBorder(nil, nil,
			container.NewHBox(
				newMinSize(left, fyne.NewSize(72, rowHeight)),
				newMinSize(gutterBox, fyne.NewSize(14, rowHeight)),
			),
			nil,
			r.code,
		),
	)
	r.ExtendBaseWidget(r)
	return r
}

func (r *unifiedRowWidget) Set(ln diff.Line) {
	bg, fg, prefix, oldS, newS := unifiedStyle(ln)
	r.bg.FillColor = bg
	r.oldNo.Text = oldS
	r.newNo.Text = newS
	r.gutter.Text = prefix
	r.gutter.Color = fg
	r.code.Text = ln.Text
	r.code.Color = fg
	r.bg.Refresh()
	r.oldNo.Refresh()
	r.newNo.Refresh()
	r.gutter.Refresh()
	r.code.Refresh()
}

func (r *unifiedRowWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.content)
}

func (r *unifiedRowWidget) MinSize() fyne.Size {
	return fyne.NewSize(200, rowHeight)
}

// --- split row ---

type splitRowWidget struct {
	widget.BaseWidget
	leftBG, rightBG     *canvas.Rectangle
	leftNo, rightNo     *canvas.Text
	leftMark, rightMark *canvas.Text
	leftCode, rightCode *canvas.Text
	content             *fyne.Container
}

func newSplitRowWidget() *splitRowWidget {
	r := &splitRowWidget{
		leftBG:    canvas.NewRectangle(colorCtxBg),
		rightBG:   canvas.NewRectangle(colorCtxBg),
		leftNo:    canvas.NewText("", colorLineNoFg),
		rightNo:   canvas.NewText("", colorLineNoFg),
		leftMark:  canvas.NewText("", theme.ForegroundColor()),
		rightMark: canvas.NewText("", theme.ForegroundColor()),
		leftCode:  canvas.NewText("", theme.ForegroundColor()),
		rightCode: canvas.NewText("", theme.ForegroundColor()),
	}
	for _, t := range []*canvas.Text{r.leftNo, r.rightNo} {
		t.TextSize = 11
		t.TextStyle = fyne.TextStyle{Monospace: true}
		t.Alignment = fyne.TextAlignTrailing
	}
	for _, t := range []*canvas.Text{r.leftMark, r.rightMark, r.leftCode, r.rightCode} {
		t.TextSize = 12
		t.TextStyle = fyne.TextStyle{Monospace: true}
	}
	r.leftMark.Alignment = fyne.TextAlignCenter
	r.rightMark.Alignment = fyne.TextAlignCenter

	left := halfPane(r.leftBG, r.leftNo, r.leftMark, r.leftCode)
	right := halfPane(r.rightBG, r.rightNo, r.rightMark, r.rightCode)
	r.content = container.New(layout.NewGridLayout(2), left, right)
	r.ExtendBaseWidget(r)
	return r
}

func halfPane(bg *canvas.Rectangle, num, mark, code *canvas.Text) fyne.CanvasObject {
	numBg := canvas.NewRectangle(colorLineNoBg)
	numBox := container.NewStack(numBg, num)
	// Do not reuse `bg` in markBox — a CanvasObject may only have one parent.
	markBox := container.NewCenter(mark)
	return container.NewStack(bg,
		container.NewBorder(nil, nil,
			container.NewHBox(
				newMinSize(numBox, fyne.NewSize(40, rowHeight)),
				newMinSize(markBox, fyne.NewSize(12, rowHeight)),
			),
			nil,
			code,
		),
	)
}

func (r *splitRowWidget) Set(row diff.SplitRow) {
	lbg, lfg := splitStyle(row.OldKind)
	rbg, rfg := splitStyle(row.NewKind)
	r.leftBG.FillColor = lbg
	r.rightBG.FillColor = rbg

	r.leftNo.Text = lineNo(row.OldNo)
	r.rightNo.Text = lineNo(row.NewNo)
	r.leftMark.Text = markFor(row.OldKind)
	r.rightMark.Text = markFor(row.NewKind)
	r.leftMark.Color = lfg
	r.rightMark.Color = rfg
	r.leftCode.Text = row.OldText
	r.rightCode.Text = row.NewText
	r.leftCode.Color = lfg
	r.rightCode.Color = rfg

	r.leftBG.Refresh()
	r.rightBG.Refresh()
	r.leftNo.Refresh()
	r.rightNo.Refresh()
	r.leftMark.Refresh()
	r.rightMark.Refresh()
	r.leftCode.Refresh()
	r.rightCode.Refresh()
}

func (r *splitRowWidget) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.content)
}

func (r *splitRowWidget) MinSize() fyne.Size {
	return fyne.NewSize(400, rowHeight)
}

func lineNo(n int) string {
	if n <= 0 {
		return ""
	}
	return strconv.Itoa(n)
}

func markFor(k diff.Kind) string {
	switch k {
	case diff.KindAdd:
		return "+"
	case diff.KindDel:
		return "-"
	default:
		return ""
	}
}

func unifiedStyle(ln diff.Line) (bg, fg color.Color, prefix, oldS, newS string) {
	bg, fg = colorCtxBg, theme.ForegroundColor()
	prefix = " "
	switch ln.Kind {
	case diff.KindAdd:
		bg, fg, prefix = colorAddBg, colorAddFg, "+"
	case diff.KindDel:
		bg, fg, prefix = colorDelBg, colorDelFg, "-"
	case diff.KindHunk:
		bg, fg, prefix = colorHunkBg, colorHunkFg, " "
		return bg, fg, prefix, "", ""
	case diff.KindMeta:
		fg = colorLineNoFg
	}
	if ln.OldNo > 0 {
		oldS = strconv.Itoa(ln.OldNo)
	}
	if ln.NewNo > 0 {
		newS = strconv.Itoa(ln.NewNo)
	}
	return
}

func splitStyle(kind diff.Kind) (bg, fg color.Color) {
	bg, fg = colorCtxBg, theme.ForegroundColor()
	switch kind {
	case diff.KindAdd:
		return colorAddBg, colorAddFg
	case diff.KindDel:
		return colorDelBg, colorDelFg
	case diff.KindHunk:
		return colorHunkBg, colorHunkFg
	}
	return
}

// minSize wraps an object with a forced minimum size.
type minSize struct {
	widget.BaseWidget
	obj fyne.CanvasObject
	min fyne.Size
}

func newMinSize(obj fyne.CanvasObject, min fyne.Size) *minSize {
	m := &minSize{obj: obj, min: min}
	m.ExtendBaseWidget(m)
	return m
}

func (m *minSize) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(m.obj)
}

func (m *minSize) MinSize() fyne.Size {
	return m.min
}

func (m *minSize) Resize(s fyne.Size) {
	m.BaseWidget.Resize(s)
	m.obj.Resize(s)
}

// FormatChurn formats +/- line counts.
func FormatChurn(add, del int, binary bool) string {
	if binary {
		return "binary"
	}
	return fmt.Sprintf("+%d −%d", add, del)
}
