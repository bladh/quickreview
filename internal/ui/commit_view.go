package ui

import (
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/bladh/quickreview/internal/git"
)

// CommitPane shows commit metadata with copy buttons, embedded beside the diff.
type CommitPane struct {
	widget.BaseWidget

	subject *widget.Label
	body    *widget.Label
	empty   *widget.Label

	shortHash *copyField
	fullHash  *copyField
	branch    *copyField
	author    *copyField
	email     *copyField
	date      *copyField
	subjectF  *copyField

	scroll *container.Scroll
	stack  *fyne.Container

	clipboard func() fyne.Clipboard
	onCopied  func(string)
}

type copyField struct {
	label  *widget.Label
	value  *widget.Label
	button *widget.Button
	box    *fyne.Container
	text   string
}

func newCopyField(title string, clipboard func() fyne.Clipboard, onCopied func(string)) *copyField {
	f := &copyField{
		label: widget.NewLabelWithStyle(title, fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		value: widget.NewLabel(""),
	}
	f.value.Wrapping = fyne.TextWrapWord
	f.value.TextStyle = fyne.TextStyle{Monospace: true}
	f.button = widget.NewButtonWithIcon("Copy", theme.ContentCopyIcon(), func() {
		if f.text == "" || clipboard == nil {
			return
		}
		if cb := clipboard(); cb != nil {
			cb.SetContent(f.text)
			if onCopied != nil {
				onCopied(title)
			}
		}
	})
	f.button.Importance = widget.LowImportance
	f.box = container.NewBorder(nil, nil, nil, f.button,
		container.NewVBox(f.label, f.value),
	)
	return f
}

func (f *copyField) Set(text string) {
	f.text = text
	if text == "" {
		f.value.SetText("—")
		f.button.Disable()
	} else {
		f.value.SetText(text)
		f.button.Enable()
	}
}

func NewCommitPane(clipboard func() fyne.Clipboard, onCopied func(string)) *CommitPane {
	p := &CommitPane{
		clipboard: clipboard,
		onCopied:  onCopied,
		subject:   widget.NewLabel("Select a commit to view its message"),
		body:      widget.NewLabel(""),
		empty:     widget.NewLabel("Select a commit to view its message"),
	}
	p.ExtendBaseWidget(p)

	p.subject.TextStyle = fyne.TextStyle{Bold: true}
	p.subject.Wrapping = fyne.TextWrapWord
	p.body.Wrapping = fyne.TextWrapWord
	p.body.TextStyle = fyne.TextStyle{Monospace: true}
	p.empty.Alignment = fyne.TextAlignCenter

	p.shortHash = newCopyField("Short SHA", clipboard, onCopied)
	p.fullHash = newCopyField("Full SHA-1", clipboard, onCopied)
	p.branch = newCopyField("Branch", clipboard, onCopied)
	p.author = newCopyField("Author", clipboard, onCopied)
	p.email = newCopyField("Email", clipboard, onCopied)
	p.date = newCopyField("Date", clipboard, onCopied)
	p.subjectF = newCopyField("Subject", clipboard, onCopied)

	fields := container.NewVBox(
		p.subject,
		widget.NewSeparator(),
		p.shortHash.box,
		p.fullHash.box,
		p.branch.box,
		p.author.box,
		p.email.box,
		p.date.box,
		p.subjectF.box,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Body", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		p.body,
	)
	p.scroll = container.NewVScroll(container.NewPadded(fields))
	p.stack = container.NewStack(p.empty, p.scroll)
	p.showEmpty()
	return p
}

func (p *CommitPane) CreateRenderer() fyne.WidgetRenderer {
	return widget.NewSimpleRenderer(p.stack)
}

func (p *CommitPane) showEmpty() {
	p.empty.Show()
	p.scroll.Hide()
}

func (p *CommitPane) Clear() {
	p.showEmpty()
	p.subject.SetText("Select a commit to view its message")
	p.body.SetText("")
	for _, f := range []*copyField{p.shortHash, p.fullHash, p.branch, p.author, p.email, p.date, p.subjectF} {
		f.Set("")
	}
}

func (p *CommitPane) Set(c git.Commit, branch string) {
	if c.Uncommitted {
		p.SetUncommitted(branch)
		return
	}
	p.empty.Hide()
	p.scroll.Show()

	p.subject.SetText(c.Subject)
	body := strings.TrimSpace(c.Body)
	if body == "" {
		body = "(no commit body)"
	}
	p.body.SetText(body)

	p.shortHash.Set(c.ShortHash)
	p.fullHash.Set(c.Hash)
	p.branch.Set(branch)
	p.author.Set(c.Author)
	p.email.Set(c.Email)
	p.date.Set(formatDate(c.Date))
	p.subjectF.Set(c.Subject)
	p.scroll.Offset = fyne.NewPos(0, 0)
	p.scroll.Refresh()
}

// SetUncommitted shows the working-tree row, which has no commit metadata.
func (p *CommitPane) SetUncommitted(branch string) {
	p.empty.Hide()
	p.scroll.Show()

	p.subject.SetText(uncommittedChangesLabel)
	p.body.SetText("Staged, unstaged, and untracked changes in the working tree, compared with HEAD.")
	p.shortHash.Set("")
	p.fullHash.Set("")
	p.branch.Set(branch)
	p.author.Set("")
	p.email.Set("")
	p.date.Set("")
	p.subjectF.Set(uncommittedChangesLabel)
	p.scroll.Offset = fyne.NewPos(0, 0)
	p.scroll.Refresh()
}

func (a *App) showCommitTab() {
	if a.detailTabs == nil {
		return
	}
	a.detailTabs.SelectIndex(0)
}

func (a *App) showDiffTab() {
	if a.detailTabs == nil {
		return
	}
	a.detailTabs.SelectIndex(1)
}

func (a *App) updateCommitPane() {
	if a.commitPane == nil {
		return
	}
	if a.selectedCommit < 0 || a.selectedCommit >= len(a.commits) {
		a.commitPane.Clear()
		return
	}
	a.commitPane.Set(a.commits[a.selectedCommit], a.selectedBranch)
}

func (a *App) copyStatus(field string) {
	a.setStatus("Copied %s to clipboard", strings.ToLower(field))
}
