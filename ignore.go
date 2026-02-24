package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type ignorer struct {
	alwaysExclude map[string]bool
	patterns      map[string][]ignorePattern
}

type ignorePattern struct {
	pattern  string
	negated  bool
	dirOnly  bool
	anchored bool
}

func newIgnorer(alwaysExclude []string, sourceRoot string) *ignorer {
	ae := make(map[string]bool)
	for _, e := range alwaysExclude {
		ae[e] = true
	}
	ig := &ignorer{
		alwaysExclude: ae,
		patterns:      make(map[string][]ignorePattern),
	}
	ig.preload(sourceRoot)
	return ig
}

func (ig *ignorer) preload(root string) {
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := filepath.Base(path)
			if ig.alwaysExclude[name] && path != root {
				return filepath.SkipDir
			}
			rel, _ := filepath.Rel(root, path)
			ig.loadGitignore(path, rel)
			return nil
		}
		return nil
	})
}

func (ig *ignorer) isExcluded(relPath string, isDir bool) bool {
	name := filepath.Base(relPath)
	if ig.alwaysExclude[name] {
		return true
	}

	dirParts := strings.Split(filepath.Dir(relPath), string(os.PathSeparator))

	dirs := []string{"."}
	current := ""
	for _, part := range dirParts {
		if part == "." {
			continue
		}
		if current == "" {
			current = part
		} else {
			current = current + string(os.PathSeparator) + part
		}
		dirs = append(dirs, current)
	}

	excluded := false
	for _, dir := range dirs {
		patterns, ok := ig.patterns[dir]
		if !ok {
			continue
		}
		for _, p := range patterns {
			if p.dirOnly && !isDir {
				continue
			}
			if matchPattern(p, relPath, name) {
				if p.negated {
					excluded = false
				} else {
					excluded = true
				}
			}
		}
	}

	return excluded
}

func (ig *ignorer) loadGitignore(absDir string, relDir string) {
	gitignorePath := filepath.Join(absDir, ".gitignore")
	f, err := os.Open(gitignorePath)
	if err != nil {
		return
	}
	defer f.Close()

	var patterns []ignorePattern
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		p := ignorePattern{}

		if strings.HasPrefix(line, "!") {
			p.negated = true
			line = line[1:]
		}

		if strings.HasSuffix(line, "/") {
			p.dirOnly = true
			line = strings.TrimSuffix(line, "/")
		}

		if strings.Contains(line, "/") {
			p.anchored = true
		}

		p.pattern = line
		patterns = append(patterns, p)
	}

	if len(patterns) > 0 {
		ig.patterns[relDir] = patterns
	}
}

func matchPattern(p ignorePattern, relPath string, baseName string) bool {
	pattern := p.pattern

	if strings.Contains(pattern, "**") {
		return matchDoublestar(pattern, relPath)
	}

	if p.anchored {
		matched, _ := filepath.Match(pattern, relPath)
		return matched
	}

	matched, _ := filepath.Match(pattern, baseName)
	if matched {
		return true
	}

	parts := strings.Split(relPath, string(os.PathSeparator))
	for i := range parts {
		sub := strings.Join(parts[i:], string(os.PathSeparator))
		matched, _ = filepath.Match(pattern, sub)
		if matched {
			return true
		}
	}

	return false
}

func matchDoublestar(pattern string, path string) bool {
	if pattern == "**" {
		return true
	}

	if strings.HasPrefix(pattern, "**/") {
		rest := pattern[3:]
		if matched, _ := filepath.Match(rest, path); matched {
			return true
		}
		parts := strings.Split(path, string(os.PathSeparator))
		for i := range parts {
			sub := strings.Join(parts[i:], string(os.PathSeparator))
			if matched, _ := filepath.Match(rest, sub); matched {
				return true
			}
		}
		return false
	}

	if strings.HasSuffix(pattern, "/**") {
		prefix := pattern[:len(pattern)-3]
		return strings.HasPrefix(path, prefix+string(os.PathSeparator)) || path == prefix
	}

	idx := strings.Index(pattern, "/**/")
	if idx >= 0 {
		prefix := pattern[:idx]
		suffix := pattern[idx+4:]
		if !strings.HasPrefix(path, prefix+string(os.PathSeparator)) && path != prefix {
			return false
		}
		rest := strings.TrimPrefix(path, prefix+string(os.PathSeparator))
		if matched, _ := filepath.Match(suffix, rest); matched {
			return true
		}
		parts := strings.Split(rest, string(os.PathSeparator))
		for i := range parts {
			sub := strings.Join(parts[i:], string(os.PathSeparator))
			if matched, _ := filepath.Match(suffix, sub); matched {
				return true
			}
		}
	}

	return false
}
