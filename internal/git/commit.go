package git

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

// Commit is a parsed git log entry.
type Commit struct {
	Hash      string
	ShortHash string
	Author    string
	Email     string
	Date      time.Time
	Subject   string
	Body      string
}

const logFormat = "%H%x1f%h%x1f%an%x1f%ae%x1f%aI%x1f%s%x1f%b%x1e"

// Log returns commits reachable from ref, optionally limited.
func (r *Repo) Log(ref string, limit int) ([]Commit, error) {
	if ref == "" {
		ref = "HEAD"
	}
	args := []string{"log", "--format=" + logFormat}
	if limit > 0 {
		args = append(args, "-n", strconv.Itoa(limit))
	}
	args = append(args, ref)

	out, err := r.run(args...)
	if err != nil {
		return nil, err
	}
	return parseLog(out)
}

// LogRange returns commits in base..head (exclusive base, inclusive head).
func (r *Repo) LogRange(base, head string, limit int) ([]Commit, error) {
	if base == "" || head == "" {
		return nil, fmt.Errorf("base and head are required")
	}
	args := []string{"log", "--format=" + logFormat}
	if limit > 0 {
		args = append(args, "-n", strconv.Itoa(limit))
	}
	args = append(args, base+".."+head)

	out, err := r.run(args...)
	if err != nil {
		return nil, err
	}
	return parseLog(out)
}

func parseLog(out string) ([]Commit, error) {
	var commits []Commit
	for _, entry := range strings.Split(out, "\x1e") {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		parts := strings.SplitN(entry, "\x1f", 7)
		if len(parts) < 6 {
			continue
		}
		ts, err := time.Parse(time.RFC3339, parts[4])
		if err != nil {
			ts = time.Time{}
		}
		body := ""
		if len(parts) > 6 {
			body = strings.TrimSpace(parts[6])
		}
		commits = append(commits, Commit{
			Hash:      parts[0],
			ShortHash: parts[1],
			Author:    parts[2],
			Email:     parts[3],
			Date:      ts,
			Subject:   parts[5],
			Body:      body,
		})
	}
	return commits, nil
}

// ShowParent returns the first parent hash of commit, or empty if root.
func (r *Repo) ShowParent(hash string) (string, error) {
	out, err := r.run("rev-list", "--parents", "-n", "1", hash)
	if err != nil {
		return "", err
	}
	fields := strings.Fields(strings.TrimSpace(out))
	if len(fields) < 2 {
		return "", nil
	}
	return fields[1], nil
}
