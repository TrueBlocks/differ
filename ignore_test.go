package main

import (
	"testing"
)

func TestMatchPattern_Simple(t *testing.T) {
	tests := []struct {
		name     string
		pattern  ignorePattern
		relPath  string
		baseName string
		want     bool
	}{
		{
			"glob matches basename",
			ignorePattern{pattern: "*.log"},
			"logs/app.log", "app.log",
			true,
		},
		{
			"glob does not match",
			ignorePattern{pattern: "*.log"},
			"logs/app.txt", "app.txt",
			false,
		},
		{
			"exact basename",
			ignorePattern{pattern: ".DS_Store"},
			"path/to/.DS_Store", ".DS_Store",
			true,
		},
		{
			"anchored pattern matches full path",
			ignorePattern{pattern: "build/output", anchored: true},
			"build/output", "output",
			true,
		},
		{
			"anchored pattern does not match nested",
			ignorePattern{pattern: "build/output", anchored: true},
			"src/build/output", "output",
			false,
		},
		{
			"unanchored matches at any depth",
			ignorePattern{pattern: "*.o"},
			"src/deep/file.o", "file.o",
			true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchPattern(tt.pattern, tt.relPath, tt.baseName)
			if got != tt.want {
				t.Errorf("matchPattern(%+v, %q, %q) = %v, want %v",
					tt.pattern, tt.relPath, tt.baseName, got, tt.want)
			}
		})
	}
}

func TestMatchDoublestar(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		path    string
		want    bool
	}{
		{"bare **", "**", "anything/at/all", true},
		{"**/file matches top", "**/Makefile", "Makefile", true},
		{"**/file matches nested", "**/Makefile", "src/Makefile", true},
		{"**/file matches deep", "**/Makefile", "a/b/c/Makefile", true},
		{"**/file no match", "**/Makefile", "src/main.go", false},
		{"dir/**", "vendor/**", "vendor/pkg/mod", true},
		{"dir/** exact", "vendor/**", "vendor", true},
		{"dir/** no match", "vendor/**", "src/vendor", false},
		{"middle /**/", "src/**/test.go", "src/pkg/test.go", true},
		{"middle /**/ deep", "src/**/test.go", "src/a/b/test.go", true},
		{"middle /**/ no match", "src/**/test.go", "lib/test.go", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchDoublestar(tt.pattern, tt.path)
			if got != tt.want {
				t.Errorf("matchDoublestar(%q, %q) = %v, want %v",
					tt.pattern, tt.path, got, tt.want)
			}
		})
	}
}

func TestIgnorer_IsExcluded(t *testing.T) {
	ig := newIgnorer([]string{".git", "node_modules"}, "/nonexistent")

	tests := []struct {
		name    string
		relPath string
		isDir   bool
		want    bool
	}{
		{"always excluded dir", ".git", true, true},
		{"always excluded dir nested", "node_modules", true, true},
		{"normal file", "src/main.go", false, false},
		{"normal dir", "src", true, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ig.isExcluded(tt.relPath, tt.isDir)
			if got != tt.want {
				t.Errorf("isExcluded(%q, %v) = %v, want %v",
					tt.relPath, tt.isDir, got, tt.want)
			}
		})
	}
}
