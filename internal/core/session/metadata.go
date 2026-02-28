package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// workspaceYAML mirrors the fields in workspace.yaml that we care about.
type workspaceYAML struct {
	ID         string    `yaml:"id"`
	CWD        string    `yaml:"cwd"`
	GitRoot    string    `yaml:"git_root"`
	Repository string    `yaml:"repository"`
	Branch     string    `yaml:"branch"`
	Summary    string    `yaml:"summary"`
	CreatedAt  time.Time `yaml:"created_at"`
	UpdatedAt  time.Time `yaml:"updated_at"`
}

// EnrichMetadata reads workspace.yaml and events.jsonl from s.RootPath and
// populates the remaining fields of s in place. It returns any non-fatal
// parsing errors via s.MetadataErr so the caller can display a warning.
func EnrichMetadata(s *Session) {
	s.MetadataErr = enrichFromWorkspace(s)
	s.EventCount = countEvents(filepath.Join(s.RootPath, "events.jsonl"))
	s.SizeBytes = dirSize(s.RootPath)
}

func enrichFromWorkspace(s *Session) error {
	path := filepath.Join(s.RootPath, "workspace.yaml")
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// No workspace.yaml — use folder mtime as UpdatedAt fallback.
			if info, statErr := os.Stat(s.RootPath); statErr == nil {
				s.UpdatedAt = info.ModTime()
			}
			return fmt.Errorf("workspace.yaml not found")
		}
		return fmt.Errorf("read workspace.yaml: %w", err)
	}

	var w workspaceYAML
	if err := yaml.Unmarshal(data, &w); err != nil {
		// Corrupt YAML — fall back to folder mtime.
		if info, statErr := os.Stat(s.RootPath); statErr == nil {
			s.UpdatedAt = info.ModTime()
		}
		return fmt.Errorf("parse workspace.yaml: %w", err)
	}

	s.CWD = w.CWD
	s.Repository = w.Repository
	s.Branch = w.Branch
	s.Summary = w.Summary
	s.CreatedAt = w.CreatedAt
	s.UpdatedAt = w.UpdatedAt

	// If UpdatedAt is zero (field missing), fall back to mtime.
	if s.UpdatedAt.IsZero() {
		if info, statErr := os.Stat(s.RootPath); statErr == nil {
			s.UpdatedAt = info.ModTime()
		}
	}
	return nil
}

// countEvents counts the number of valid JSON lines in events.jsonl.
// Lines that are empty or contain invalid JSON are skipped.
// Returns -1 if the file cannot be opened.
func countEvents(path string) int {
	f, err := os.Open(path)
	if err != nil {
		return -1
	}
	defer f.Close()

	count := 0
	sc := bufio.NewScanner(f)
	// Increase buffer capacity for sessions with large event payloads.
	sc.Buffer(make([]byte, 64*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) > 0 && json.Valid(line) {
			count++
		}
	}
	return count
}

// dirSize returns the total byte size of all regular files under dir.
func dirSize(dir string) int64 {
	var total int64
	_ = filepath.WalkDir(dir, func(_ string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		info, err := d.Info()
		if err != nil {
			return nil
		}
		total += info.Size()
		return nil
	})
	return total
}
