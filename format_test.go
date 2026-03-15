package main

import (
	"testing"
)

func TestTruncatePath(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		maxLen int
		want   string
	}{
		{"short path unchanged", "foo/bar", 20, "foo/bar"},
		{"exact length", "abcde", 5, "abcde"},
		{"truncated with ellipsis", "very/long/path/to/file.txt", 15, ".../to/file.txt"},
		{"maxLen <= 3 no ellipsis", "abcdefgh", 3, "abc"},
		{"maxLen 1", "abcdefgh", 1, "a"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncatePath(tt.path, tt.maxLen)
			if got != tt.want {
				t.Errorf("truncatePath(%q, %d) = %q, want %q", tt.path, tt.maxLen, got, tt.want)
			}
		})
	}
}

func TestSplitGroupAndFile(t *testing.T) {
	tests := []struct {
		name      string
		relPath   string
		wantGroup string
		wantFile  string
	}{
		{"nested path", "docs/readme.md", "docs", "readme.md"},
		{"top-level file", "readme.md", ".", "readme.md"},
		{"deep path", "a/b/c/d.txt", "a", "b/c/d.txt"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group, file := splitGroupAndFile(tt.relPath)
			if group != tt.wantGroup || file != tt.wantFile {
				t.Errorf("splitGroupAndFile(%q) = (%q, %q), want (%q, %q)",
					tt.relPath, group, file, tt.wantGroup, tt.wantFile)
			}
		})
	}
}

func TestShortDetails(t *testing.T) {
	tests := []struct {
		name    string
		details []string
		want    []string
	}{
		{"size without docx", []string{"size: 10 vs 20"}, []string{"size"}},
		{"size with docx suppressed", []string{"size: 10 vs 20", "docx:text"}, []string{"docx:text"}},
		{"mode", []string{"mode: 0644 vs 0755"}, []string{"mode"}},
		{"hash", []string{"hash: abc vs def"}, []string{"hash"}},
		{"modified", []string{"modified: 2026-01-01 vs 2026-01-02"}, []string{"date"}},
		{"mixed", []string{"mode: 0644 vs 0755", "size: 10 vs 20"}, []string{"mode", "size"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shortDetails(tt.details)
			if len(got) != len(tt.want) {
				t.Fatalf("shortDetails(%v) returned %d items, want %d: %v", tt.details, len(got), len(tt.want), got)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("shortDetails(%v)[%d] = %q, want %q", tt.details, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestColorize_NoTTY(t *testing.T) {
	// When stdout is not a TTY (as in tests), colorize should return the string unchanged.
	got := colorize("31", "hello")
	if got != "hello" {
		t.Errorf("colorize in non-TTY should return plain string, got %q", got)
	}
}
