package ui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"github.com/bladh/quickreview/internal/git"
)

func (a *App) requireRepo() bool {
	if a.repo != nil {
		return true
	}
	dialog.ShowInformation("Repository", "Open a repository first (File → Open Repository).", a.win)
	return false
}

func (a *App) runGitOp(title string, fn func() (string, error)) {
	if !a.requireRepo() {
		return
	}
	a.setStatus("%s…", title)
	go func() {
		out, err := fn()
		fyne.Do(func() {
			if err != nil {
				dialog.ShowError(err, a.win)
				a.setStatus("%s failed", title)
				return
			}
			msg := strings.TrimSpace(out)
			if msg == "" {
				msg = title + " completed"
			} else if len(msg) > 200 {
				msg = msg[:200] + "…"
			}
			a.reloadRepo()
			a.setStatus("%s", msg)
		})
	}()
}

func (a *App) menuFetch() {
	a.runGitOp("Fetch", func() (string, error) {
		out, err := a.repo.Fetch()
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(out) == "" {
			return "Fetch completed (already up to date)", nil
		}
		return "Fetch completed", nil
	})
}

func (a *App) menuPull() {
	a.runGitOp("Pull", func() (string, error) {
		out, err := a.repo.Pull()
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(out) == "" {
			return "Pull completed", nil
		}
		return "Pull completed", nil
	})
}

func (a *App) menuPush() {
	a.runGitOp("Push", func() (string, error) {
		_, err := a.repo.Push()
		if err != nil {
			return "", err
		}
		return "Push completed", nil
	})
}

func (a *App) menuCheckout() {
	if !a.requireRepo() {
		return
	}
	locals := git.LocalBranches(a.branches)
	if len(locals) == 0 {
		dialog.ShowInformation("Checkout", "No local branches available.", a.win)
		return
	}
	names := make([]string, 0, len(locals))
	selected := ""
	for _, b := range locals {
		names = append(names, b.Name)
		if b.Current {
			selected = b.Name
		}
	}
	sel := widget.NewSelect(names, nil)
	if selected != "" {
		sel.SetSelected(selected)
	} else if len(names) > 0 {
		sel.SetSelected(names[0])
	}

	form := container.NewVBox(
		widget.NewLabel("Select a local branch to check out:"),
		sel,
	)
	dialog.ShowCustomConfirm("Checkout Branch", "Checkout", "Cancel", form, func(ok bool) {
		if !ok || sel.Selected == "" {
			return
		}
		branch := sel.Selected
		a.runGitOp("Checkout "+branch, func() (string, error) {
			_, err := a.repo.Checkout(branch)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Checked out %s", branch), nil
		})
	}, a.win)
}

func (a *App) menuCreateBranch() {
	if !a.requireRepo() {
		return
	}
	nameEntry := widget.NewEntry()
	nameEntry.SetPlaceHolder("new-branch-name")
	startEntry := widget.NewEntry()
	startEntry.SetText("HEAD")
	startEntry.SetPlaceHolder("start point (branch, tag, or commit)")

	form := container.NewVBox(
		widget.NewLabel("Branch name"),
		nameEntry,
		widget.NewLabel("Start point"),
		startEntry,
	)
	dialog.ShowCustomConfirm("Create Branch", "Create", "Cancel", form, func(ok bool) {
		if !ok {
			return
		}
		name := strings.TrimSpace(nameEntry.Text)
		start := strings.TrimSpace(startEntry.Text)
		if name == "" {
			dialog.ShowInformation("Create Branch", "Enter a branch name.", a.win)
			return
		}
		a.runGitOp("Create "+name, func() (string, error) {
			_, err := a.repo.CreateBranch(name, start)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Created and checked out %s", name), nil
		})
	}, a.win)
}

func (a *App) menuViewRemoteBranches() {
	if !a.requireRepo() {
		return
	}
	remotes := git.RemoteBranches(a.branches)
	git.SortBranches(remotes, a.prefs.branchSortMode())
	if len(remotes) == 0 {
		dialog.ShowInformation("Remote Branches", "No remote-tracking branches found.\nTry Repository → Fetch first.", a.win)
		return
	}

	names := make([]string, len(remotes))
	for i, b := range remotes {
		label := b.Name
		if !b.CommitTime.IsZero() {
			label = fmt.Sprintf("%s  (%s)", b.Name, b.CommitTime.Local().Format("2006-01-02"))
		}
		names[i] = label
	}

	list := widget.NewList(
		func() int { return len(remotes) },
		func() fyne.CanvasObject { return widget.NewLabel("remote/branch") },
		func(id widget.ListItemID, obj fyne.CanvasObject) {
			if id >= len(names) {
				return
			}
			obj.(*widget.Label).SetText(names[id])
		},
	)
	selected := -1
	list.OnSelected = func(id widget.ListItemID) { selected = id }

	hint := widget.NewLabel("Select a remote branch, then Checkout to create a local tracking branch.")
	hint.Wrapping = fyne.TextWrapWord

	body := container.NewBorder(hint, nil, nil, nil, list)
	body.Resize(fyne.NewSize(480, 360))

	var d dialog.Dialog
	checkoutBtn := widget.NewButton("Checkout…", func() {
		if selected < 0 || selected >= len(remotes) {
			dialog.ShowInformation("Checkout", "Select a remote branch first.", a.win)
			return
		}
		remote := remotes[selected].Name
		a.runGitOp("Checkout "+remote, func() (string, error) {
			_, err := a.repo.CheckoutRemote(remote)
			if err != nil {
				return "", err
			}
			return fmt.Sprintf("Checked out %s", remote), nil
		})
		if d != nil {
			d.Hide()
		}
	})
	checkoutBtn.Importance = widget.HighImportance
	closeBtn := widget.NewButton("Close", func() {
		if d != nil {
			d.Hide()
		}
	})

	content := container.NewBorder(nil, container.NewHBox(checkoutBtn, closeBtn), nil, nil, body)
	d = dialog.NewCustomWithoutButtons("Remote Branches", content, a.win)
	d.Resize(fyne.NewSize(520, 440))
	d.Show()
}

func (a *App) repositoryMenu() *fyne.Menu {
	return fyne.NewMenu("Repository",
		fyne.NewMenuItem("Fetch", a.menuFetch),
		fyne.NewMenuItem("Pull", a.menuPull),
		fyne.NewMenuItem("Push", a.menuPush),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("Checkout…", a.menuCheckout),
		fyne.NewMenuItem("Create Branch…", a.menuCreateBranch),
		fyne.NewMenuItemSeparator(),
		fyne.NewMenuItem("View Remote Branches…", a.menuViewRemoteBranches),
	)
}
