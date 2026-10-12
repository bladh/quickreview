# QuickReview

A cross-platform desktop git review app built with [Go](https://go.dev) and [Fyne](https://fyne.io). It shells out to the **system `git`** binary — no embedded git library — so branch history, diffs, and rename detection match the tools you already use.

**Repository:** [github.com/bladh/quickreview](https://github.com/bladh/quickreview) · **Website:** [bladh.ninja](https://bladh.ninja) · **License:** [MIT](LICENSE)

Inspired by GitHub pull-request review and clients like SmartGit: browse branches and commits, compare refs (including three-dot / merge-base diffs), walk the changed-file list, and inspect each file in **unified** or **split** view.

## Features

- Open any local git working tree
- Branch list (local + remotes), with the current branch highlighted
- Commit log for the selected branch (author, date, subject)
- **Compare** tab: base vs head, with optional **three-dot** mode (GitHub PR style)
- Changed-file list with status (`A`/`M`/`D`/`R`) and `+/-` churn
- Unified and side-by-side split diff rendering
- **Working Tree** review (staged + unstaged vs `HEAD`)
- Prev/next file navigation in the toolbar
- Works on **Linux**, **Windows**, and **macOS** (Fyne + system git)

## Requirements

- Go 1.21+
- Git installed and on `PATH`
- Fyne system dependencies for your OS ([Fyne getting started](https://docs.fyne.io/started/))

### Linux packages (Debian/Ubuntu)

```bash
sudo apt install gcc libgl1-mesa-dev xorg-dev
```

## Build & run

```bash
go mod tidy
go run .                          # opens the current directory if it is a repo
go run . /path/to/repo            # open a specific repository
go build -o quickreview .
```

### Cross-compile notes

Fyne apps need OS-specific builds (CGo / graphics). Prefer the GitHub Actions builds below, or build on each target platform / use [`fyne-cross`](https://github.com/fyne-io/fyne-cross).

## CI binaries (GitHub Actions)

Workflows under `.github/workflows/` build **Linux**, **Windows**, and **macOS** (Intel + Apple Silicon) binaries.

| Trigger | What you get |
|---------|----------------|
| Push / PR to `main` or `master` | Downloadable **Actions artifacts** (zip per platform, 30-day retention) |
| Tag `v*` (e.g. `v0.1.0`) | Same artifacts **plus** a **GitHub Release** with the zips attached |
| Actions → **Build** / **Release** → Run workflow | Manual rebuild |

**Download from a workflow run:** open the Actions run → **Artifacts** at the bottom → download `quickreview-linux-amd64`, `quickreview-windows-amd64`, `quickreview-darwin-amd64`, or `quickreview-darwin-arm64`.

**Publish a release:**

```bash
git tag v0.1.0
git push origin v0.1.0
```

Then open [Releases](https://github.com/bladh/quickreview/releases) for the published zips.

## Review workflow

1. **File → Open Repository** (or pass a path on the CLI).
2. Select a **branch** → browse **commits** → pick a commit to see its file changes.
3. Or open the **Compare** tab, choose base/head (e.g. `main`…`feature`), enable three-dot for PR-style diffs, and click **Compare**.
4. Click files in the changed list; use **View → Unified Diff** / **Split Diff** as needed.
5. **File → Preferences** sets the default browse folder, **Dark** / **Light** color scheme, and branch sort order.
6. **Repository** menu supports Fetch, Pull, Push, Checkout, Create Branch, and viewing remote branches.

## Layout

The main window puts the **diff viewer full-width on top**. Below it, **Branches**, **Commits/Compare**, and **Changed files** share the remaining width (all splitters are resizable).

```
internal/git/    system-git wrappers (branches, log, diffs)
internal/diff/   unified-diff parser + split-row alignment
internal/ui/     Fyne window, lists, diff viewer, theme
```
