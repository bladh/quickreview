package git

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// FileStatus is a short status letter for a changed path.
type FileStatus string

const (
	StatusAdded    FileStatus = "A"
	StatusModified FileStatus = "M"
	StatusDeleted  FileStatus = "D"
	StatusRenamed  FileStatus = "R"
	StatusCopied   FileStatus = "C"
	StatusType     FileStatus = "T"
	StatusUnknown  FileStatus = "?"
)

// ChangedFile is one path in a diff.
type ChangedFile struct {
	Path      string
	OldPath   string
	Status    FileStatus
	Additions int
	Deletions int
	Binary    bool
}

// DiffSpec identifies what to compare.
type DiffSpec struct {
	// Mode is "commit", "range", or "working".
	Mode string
	// For commit: show that commit vs its parent.
	Commit string
	// For range: Base..Head (or Base...Head for merge-base).
	Base string
	Head string
	// ThreeDot uses merge-base comparison (GitHub PR style).
	ThreeDot bool
}

// CommitDiffSpec returns a spec that shows a single commit's changes.
func CommitDiffSpec(hash string) DiffSpec {
	return DiffSpec{Mode: "commit", Commit: hash}
}

// RangeDiffSpec compares base and head. threeDot mirrors GitHub PR diffs.
func RangeDiffSpec(base, head string, threeDot bool) DiffSpec {
	return DiffSpec{Mode: "range", Base: base, Head: head, ThreeDot: threeDot}
}

// WorkingDiffSpec compares the working tree and index to HEAD, including
// untracked files.
func WorkingDiffSpec() DiffSpec {
	return DiffSpec{Mode: "working"}
}

// HasUncommittedChanges reports whether the index or work tree differs from
// HEAD, including untracked files. The result always describes the checked-out
// branch.
func (r *Repo) HasUncommittedChanges() (bool, error) {
	out, err := r.run("status", "--porcelain")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(out) != "", nil
}

func (r *Repo) revArgs(spec DiffSpec) ([]string, error) {
	switch spec.Mode {
	case "working":
		// Staged + unstaged changes relative to HEAD (SmartGit-style working tree).
		return []string{"HEAD"}, nil
	case "range":
		if spec.Base == "" || spec.Head == "" {
			return nil, fmt.Errorf("base and head are required")
		}
		sep := ".."
		if spec.ThreeDot {
			sep = "..."
		}
		return []string{spec.Base + sep + spec.Head}, nil
	case "commit":
		if spec.Commit == "" {
			return nil, fmt.Errorf("commit is required")
		}
		parent, err := r.ShowParent(spec.Commit)
		if err != nil {
			return nil, err
		}
		if parent == "" {
			// Well-known empty tree — works for root commits on all platforms.
			return []string{"4b825dc642cb6eb9a060e54bf8d69288fbee4904", spec.Commit}, nil
		}
		return []string{parent, spec.Commit}, nil
	default:
		return nil, fmt.Errorf("unknown diff mode %q", spec.Mode)
	}
}

// ChangedFiles lists files changed by the given spec.
func (r *Repo) ChangedFiles(spec DiffSpec) ([]ChangedFile, error) {
	revs, err := r.revArgs(spec)
	if err != nil {
		return nil, err
	}

	statusArgs := append([]string{"diff", "--name-status", "--find-renames"}, revs...)
	statusOut, err := r.run(statusArgs...)
	if err != nil {
		return nil, err
	}

	statArgs := append([]string{"diff", "--numstat", "--find-renames"}, revs...)
	statOut, err := r.run(statArgs...)
	if err != nil {
		return nil, err
	}

	files, err := parseNameStatus(statusOut)
	if err != nil {
		return nil, err
	}
	applyNumstat(files, statOut)
	if spec.Mode == "working" {
		extra, err := r.untrackedFiles()
		if err != nil {
			return nil, err
		}
		files = append(files, extra...)
	}
	return files, nil
}

// FileDiff returns the unified diff text for a single path.
func (r *Repo) FileDiff(spec DiffSpec, path string) (string, error) {
	if spec.Mode == "working" {
		untracked, err := r.isUntracked(path)
		if err != nil {
			return "", err
		}
		if untracked {
			return r.diffUntracked(path)
		}
	}
	revs, err := r.revArgs(spec)
	if err != nil {
		return "", err
	}
	args := append([]string{"diff", "--find-renames", "--unified=3"}, revs...)
	args = append(args, "--", path)
	return r.run(args...)
}

// FileContentAt returns blob content at ref:path. Missing blobs return empty.
func (r *Repo) FileContentAt(ref, path string) (string, error) {
	if ref == "" {
		return "", nil
	}
	out, err := r.run("show", ref+":"+path)
	if err != nil {
		return "", err
	}
	return out, nil
}

func (r *Repo) untrackedFiles() ([]ChangedFile, error) {
	out, err := r.run("ls-files", "--others", "--exclude-standard", "-z")
	if err != nil {
		return nil, err
	}
	paths := splitNUL(out)
	files := make([]ChangedFile, 0, len(paths))
	for _, path := range paths {
		additions, binary := workingFileChurn(filepath.Join(r.Path, filepath.FromSlash(path)))
		files = append(files, ChangedFile{
			Path:      path,
			Status:    StatusAdded,
			Additions: additions,
			Binary:    binary,
		})
	}
	return files, nil
}

func (r *Repo) isUntracked(path string) (bool, error) {
	out, err := r.run("ls-files", "--others", "--exclude-standard", "-z", "--", path)
	if err != nil {
		return false, err
	}
	for _, p := range splitNUL(out) {
		if p == path {
			return true, nil
		}
	}
	return false, nil
}

func (r *Repo) diffUntracked(path string) (string, error) {
	// --no-index exits 1 when the file differs from /dev/null.
	return runAllowing(r.Path, map[int]struct{}{1: {}},
		"diff", "--no-index", "--unified=3", "--", "/dev/null", path)
}

func workingFileChurn(path string) (additions int, binary bool) {
	f, err := os.Open(path)
	if err != nil {
		return 0, false
	}
	defer f.Close()

	buf := make([]byte, 32*1024)
	sawBytes := false
	partial := false
	for {
		n, err := f.Read(buf)
		if n > 0 {
			chunk := buf[:n]
			if bytes.IndexByte(chunk, 0) >= 0 {
				return 0, true
			}
			sawBytes = true
			additions += bytes.Count(chunk, []byte{'\n'})
			partial = chunk[n-1] != '\n'
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return additions, false
		}
	}
	if sawBytes && partial {
		additions++
	}
	return additions, false
}

func splitNUL(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, "\x00")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}

func parseNameStatus(out string) ([]ChangedFile, error) {
	var files []ChangedFile
	for _, line := range splitNonEmpty(out) {
		fields := strings.Split(line, "\t")
		if len(fields) < 2 {
			continue
		}
		raw := fields[0]
		status := FileStatus(raw[:1])
		cf := ChangedFile{Status: status}
		switch {
		case len(fields) >= 3 && (status == StatusRenamed || status == StatusCopied):
			cf.OldPath = fields[1]
			cf.Path = fields[2]
		case status == StatusDeleted:
			cf.Path = fields[1]
			cf.OldPath = fields[1]
		default:
			cf.Path = fields[1]
		}
		files = append(files, cf)
	}
	return files, nil
}

func applyNumstat(files []ChangedFile, out string) {
	byPath := make(map[string]*ChangedFile, len(files))
	for i := range files {
		byPath[files[i].Path] = &files[i]
		if files[i].OldPath != "" {
			byPath[files[i].OldPath] = &files[i]
		}
	}
	for _, line := range splitNonEmpty(out) {
		fields := strings.Split(line, "\t")
		if len(fields) < 3 {
			continue
		}
		path := fields[len(fields)-1]
		cf, ok := byPath[path]
		if !ok && len(fields) >= 3 {
			// rename numstat: old\tnew sometimes — last field is new path
			cf, ok = byPath[fields[len(fields)-1]]
		}
		if !ok {
			continue
		}
		if fields[0] == "-" || fields[1] == "-" {
			cf.Binary = true
			continue
		}
		cf.Additions, _ = strconv.Atoi(fields[0])
		cf.Deletions, _ = strconv.Atoi(fields[1])
	}
}

// StatusLabel returns a short human label for a file status.
func StatusLabel(s FileStatus) string {
	switch s {
	case StatusAdded:
		return "added"
	case StatusModified:
		return "modified"
	case StatusDeleted:
		return "deleted"
	case StatusRenamed:
		return "renamed"
	case StatusCopied:
		return "copied"
	default:
		return "changed"
	}
}
