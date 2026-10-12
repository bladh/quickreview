package diff

import (
	"strings"
)

// Kind classifies a diff line.
type Kind int

const (
	KindContext Kind = iota
	KindAdd
	KindDel
	KindHeader
	KindHunk
	KindMeta
)

// Line is one visual line in a parsed unified diff.
type Line struct {
	Kind    Kind
	Text    string // without leading +/-/space for content lines
	Raw     string
	OldNo   int // 0 if not applicable
	NewNo   int // 0 if not applicable
}

// Hunk is a contiguous change region.
type Hunk struct {
	Header   string
	OldStart int
	OldCount int
	NewStart int
	NewCount int
	Lines    []Line
}

// FileDiff is a parsed unified diff for one file.
type FileDiff struct {
	OldFile string
	NewFile string
	Hunks   []Hunk
	Binary  bool
	Raw     string
}

// ParseUnified parses a unified diff (possibly multi-file) and returns the
// first file's structured representation. Empty input yields an empty FileDiff.
func ParseUnified(raw string) FileDiff {
	fd := FileDiff{Raw: raw}
	if strings.TrimSpace(raw) == "" {
		return fd
	}
	if strings.Contains(raw, "Binary files") || strings.Contains(raw, "GIT binary patch") {
		fd.Binary = true
	}

	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	var cur *Hunk
	oldNo, newNo := 0, 0

	flush := func() {
		if cur != nil {
			fd.Hunks = append(fd.Hunks, *cur)
			cur = nil
		}
	}

	for _, rawLine := range lines {
		switch {
		case strings.HasPrefix(rawLine, "diff --git"):
			flush()
		case strings.HasPrefix(rawLine, "--- "):
			fd.OldFile = trimDiffPath(strings.TrimPrefix(rawLine, "--- "))
		case strings.HasPrefix(rawLine, "+++ "):
			fd.NewFile = trimDiffPath(strings.TrimPrefix(rawLine, "+++ "))
		case strings.HasPrefix(rawLine, "@@"):
			flush()
			h := parseHunkHeader(rawLine)
			cur = &h
			oldNo = h.OldStart
			newNo = h.NewStart
		case cur == nil:
			// file-level meta (index, mode, etc.)
			continue
		case strings.HasPrefix(rawLine, "+"):
			cur.Lines = append(cur.Lines, Line{
				Kind: KindAdd, Text: rawLine[1:], Raw: rawLine, NewNo: newNo,
			})
			newNo++
		case strings.HasPrefix(rawLine, "-"):
			cur.Lines = append(cur.Lines, Line{
				Kind: KindDel, Text: rawLine[1:], Raw: rawLine, OldNo: oldNo,
			})
			oldNo++
		case strings.HasPrefix(rawLine, "\\"):
			cur.Lines = append(cur.Lines, Line{Kind: KindMeta, Text: rawLine, Raw: rawLine})
		default:
			text := rawLine
			if strings.HasPrefix(rawLine, " ") {
				text = rawLine[1:]
			}
			cur.Lines = append(cur.Lines, Line{
				Kind: KindContext, Text: text, Raw: rawLine, OldNo: oldNo, NewNo: newNo,
			})
			oldNo++
			newNo++
		}
	}
	flush()
	return fd
}

func trimDiffPath(p string) string {
	p = strings.TrimSpace(p)
	if p == "/dev/null" {
		return ""
	}
	// strip a/ or b/ prefix and optional timestamp
	if tab := strings.IndexByte(p, '\t'); tab >= 0 {
		p = p[:tab]
	}
	if len(p) >= 2 && (p[0] == 'a' || p[0] == 'b') && p[1] == '/' {
		p = p[2:]
	}
	return p
}

func parseHunkHeader(line string) Hunk {
	h := Hunk{Header: line}
	// @@ -l,s +l,s @@ optional
	start := strings.Index(line, "@@")
	end := strings.LastIndex(line, "@@")
	if start < 0 || end <= start {
		return h
	}
	body := strings.TrimSpace(line[start+2 : end])
	parts := strings.Fields(body)
	if len(parts) >= 1 {
		h.OldStart, h.OldCount = parseRange(parts[0])
	}
	if len(parts) >= 2 {
		h.NewStart, h.NewCount = parseRange(parts[1])
	}
	return h
}

func parseRange(s string) (start, count int) {
	s = strings.TrimPrefix(s, "-")
	s = strings.TrimPrefix(s, "+")
	count = 1
	if comma := strings.IndexByte(s, ','); comma >= 0 {
		fmtAtoi(s[:comma], &start)
		fmtAtoi(s[comma+1:], &count)
		return
	}
	fmtAtoi(s, &start)
	return
}

func fmtAtoi(s string, v *int) {
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	*v = n
}

// SplitSides builds old/new aligned rows for side-by-side display.
type SplitRow struct {
	OldText string
	NewText string
	OldNo   int
	NewNo   int
	OldKind Kind // context or del (or meta)
	NewKind Kind // context or add
}

// ToSplit converts hunks into paired rows (SmartGit / GitHub split style).
func (fd FileDiff) ToSplit() []SplitRow {
	var rows []SplitRow
	for _, h := range fd.Hunks {
		rows = append(rows, SplitRow{
			OldText: h.Header,
			NewText: h.Header,
			OldKind: KindHunk,
			NewKind: KindHunk,
		})

		var dels []Line
		var adds []Line
		flushPair := func() {
			max := len(dels)
			if len(adds) > max {
				max = len(adds)
			}
			for i := 0; i < max; i++ {
				row := SplitRow{}
				if i < len(dels) {
					row.OldText = dels[i].Text
					row.OldNo = dels[i].OldNo
					row.OldKind = KindDel
				}
				if i < len(adds) {
					row.NewText = adds[i].Text
					row.NewNo = adds[i].NewNo
					row.NewKind = KindAdd
				}
				rows = append(rows, row)
			}
			dels = nil
			adds = nil
		}

		for _, ln := range h.Lines {
			switch ln.Kind {
			case KindDel:
				dels = append(dels, ln)
			case KindAdd:
				adds = append(adds, ln)
			case KindContext:
				flushPair()
				rows = append(rows, SplitRow{
					OldText: ln.Text,
					NewText: ln.Text,
					OldNo:   ln.OldNo,
					NewNo:   ln.NewNo,
					OldKind: KindContext,
					NewKind: KindContext,
				})
			default:
				flushPair()
			}
		}
		flushPair()
	}
	return rows
}

// FlatLines returns all content lines in unified order (with hunk headers).
func (fd FileDiff) FlatLines() []Line {
	var out []Line
	for _, h := range fd.Hunks {
		out = append(out, Line{Kind: KindHunk, Text: h.Header, Raw: h.Header})
		out = append(out, h.Lines...)
	}
	return out
}
