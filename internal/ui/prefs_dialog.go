package ui

import (
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
)

func (a *App) showPreferences() {
	pathEntry := widget.NewEntry()
	pathEntry.SetText(a.prefs.DefaultBrowsePath)
	pathEntry.SetPlaceHolder("Folder used when opening repositories")

	browseBtn := widget.NewButtonWithIcon("Browse…", theme.FolderOpenIcon(), func() {
		d := dialog.NewFolderOpen(func(uri fyne.ListableURI, err error) {
			if err != nil {
				dialog.ShowError(err, a.win)
				return
			}
			if uri == nil {
				return
			}
			pathEntry.SetText(uri.Path())
		}, a.win)
		start := browseStartPath(pathEntry.Text, "")
		if start != "" {
			if u, err := storage.ListerForURI(storage.NewFileURI(start)); err == nil {
				d.SetLocation(u)
			}
		}
		d.Show()
		d.Resize(fyne.NewSize(800, 560))
	})

	schemeSelect := widget.NewRadioGroup([]string{"Dark", "Light"}, nil)
	switch a.prefs.ColorScheme {
	case SchemeLight:
		schemeSelect.SetSelected("Light")
	default:
		schemeSelect.SetSelected("Dark")
	}
	schemeSelect.Required = true
	schemeSelect.Horizontal = true

	sortSelect := widget.NewRadioGroup([]string{
		"Alphabetically",
		"Latest commit first",
	}, nil)
	switch a.prefs.BranchSort {
	case BranchSortRecent:
		sortSelect.SetSelected("Latest commit first")
	default:
		sortSelect.SetSelected("Alphabetically")
	}
	sortSelect.Required = true

	form := container.NewVBox(
		widget.NewLabelWithStyle("Default repository browse location", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Used as the starting folder for File → Open Repository."),
		container.NewBorder(nil, nil, nil, browseBtn, pathEntry),
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Color scheme", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Applies immediately to the main window and diff colors."),
		schemeSelect,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Branch list sorting", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("How branches are ordered in the Branches pane (current branch stays first)."),
		sortSelect,
	)

	d := dialog.NewCustomConfirm("Preferences", "Save", "Cancel", container.NewVScroll(form), func(ok bool) {
		if !ok {
			return
		}
		scheme := SchemeDark
		if schemeSelect.Selected == "Light" {
			scheme = SchemeLight
		}
		sortMode := BranchSortAlpha
		if sortSelect.Selected == "Latest commit first" {
			sortMode = BranchSortRecent
		}
		a.prefs.DefaultBrowsePath = pathEntry.Text
		a.prefs.ColorScheme = scheme
		a.prefs.BranchSort = sortMode
		a.prefs.save(a.fyneApp)
		a.applyColorScheme(scheme)
		if a.repo != nil {
			a.reloadRepo()
		}
	}, a.win)
	d.Resize(fyne.NewSize(560, 480))
	d.Show()
}

func (a *App) applyColorScheme(scheme string) {
	applyTheme(a.fyneApp, scheme)
	if a.diffView != nil {
		a.diffView.empty.Color = colorLineNoFg
		a.diffView.empty.Refresh()
		// Re-apply current diff so row colors pick up the new palette.
		if a.diffView.fd.Binary || len(a.diffView.fd.Hunks) > 0 || a.diffView.fd.Raw != "" {
			a.diffView.SetDiff(a.diffView.fd)
		}
		a.diffView.Refresh()
	}
	if a.win != nil && a.win.Content() != nil {
		a.win.Content().Refresh()
	}
}
