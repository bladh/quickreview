package ui

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/bladh/quickreview/internal/git"
)

type branchRow struct {
	widget.BaseWidget
	marker *widget.Label
	name   *widget.Label
	box    *fyne.Container
}

func newBranchRow() *branchRow {
	r := &branchRow{
		marker: widget.NewLabel("○"),
		name:   widget.NewLabel(""),
	}
	r.box = container.NewBorder(nil, nil, r.marker, nil, r.name)
	r.ExtendBaseWidget(r)
	return r
}

func (r *branchRow) Set(b git.Branch) {
	switch {
	case b.Current:
		r.marker.SetText("●")
	case b.Remote:
		r.marker.SetText("⇢")
	default:
		r.marker.SetText("○")
	}
	r.name.SetText(b.Name)
}

func (r *branchRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.box)
}

type commitRow struct {
	widget.BaseWidget
	hash *widget.Label
	subj *widget.Label
	meta *widget.Label
	box  *fyne.Container
}

func newCommitRow() *commitRow {
	r := &commitRow{
		hash: widget.NewLabel(""),
		subj: widget.NewLabel(""),
		meta: widget.NewLabel(""),
	}
	r.hash.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	r.subj.Wrapping = fyne.TextTruncate
	r.meta.TextStyle = fyne.TextStyle{Italic: true}
	r.box = container.NewBorder(nil, r.meta, r.hash, nil, r.subj)
	r.ExtendBaseWidget(r)
	return r
}

func (r *commitRow) Set(c git.Commit) {
	if c.Uncommitted {
		r.hash.SetText("*")
		r.subj.SetText(uncommittedChangesLabel)
		r.meta.SetText("working tree vs HEAD")
		return
	}
	r.hash.SetText(c.ShortHash)
	r.subj.SetText(c.Subject)
	r.meta.SetText(fmt.Sprintf("%s · %s", c.Author, formatDate(c.Date)))
}

func (r *commitRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.box)
}

func (r *commitRow) MinSize() fyne.Size {
	s := r.box.MinSize()
	if s.Height < 48 {
		s.Height = 48
	}
	return s
}

type fileRow struct {
	widget.BaseWidget
	status *widget.Label
	path   *widget.Label
	churn  *widget.Label
	box    *fyne.Container
}

func newFileRow() *fileRow {
	r := &fileRow{
		status: widget.NewLabel(""),
		path:   widget.NewLabel(""),
		churn:  widget.NewLabel(""),
	}
	r.status.TextStyle = fyne.TextStyle{Bold: true, Monospace: true}
	r.path.Wrapping = fyne.TextTruncate
	r.churn.Alignment = fyne.TextAlignTrailing
	r.box = container.NewBorder(nil, nil, r.status, r.churn, r.path)
	r.ExtendBaseWidget(r)
	return r
}

func (r *fileRow) Set(f git.ChangedFile) {
	r.status.SetText(string(f.Status))
	r.path.SetText(f.Path)
	r.churn.SetText(FormatChurn(f.Additions, f.Deletions, f.Binary))
}

func (r *fileRow) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(r.box)
}
