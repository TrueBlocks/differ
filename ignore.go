package main

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type ignorer struct {
	alwaysExclude map[string]bool
	cache         map[string][]ignorePattern
}

type ignorePattern struct {
	pattern  string
	negated  bool
	dirOnly  bool
	anchored bool
}

func newIgnorer(alwaysExclude []string) *ignorer {
	ae := make(map[string]bool)
	for _, e := range alwaysExclude {
		ae[e] = true
	}
	return &ignorer{
		alwaysExclude: ae,
		cache:         make(map[string][]ignorePattern),
	}
}

func (ig *ignorer) isExcluded(path string, root string, isDir bool) bool {
	name := filepath.Base(path)
	if ig.alwaysExclude[name] {
		return true
	}

	rel, err := filepath.Rel(root, path)
	if err != nil {
		return false
	}

	parts := strings.Split(filepath.Dir(rel), string(os.PathSeparator))
	current := root
	for _, part := range parts {
		if part == "." {
			ig.loadGitignore(current)
		} else {
			current = filepath.Join(current, part)
			ig.loadGitignore(current)
		}
	}

	ig.loadGitignore(filepath.Dir(filepath.Join(root, rel)))

	excluded := false
	current = root
	dirs := []string{root}
	for _, part := range parts {
		if part != "." {
			current = filepath.Join(current, part)
			dirs = append(dirs, current)
		}
	}

	for _, dir := range dirs {
		patterns, ok := ig.cache[dir]
		if !ok {
			continue
		}
		for _, p := range patterns {
			if p.dirOnly && !isDir {
				continue
			}
			if matchPattern(p, rel, name) {
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

func (ig *ignorer) loadGitignore(dir string) {
	if _, loaded := ig.cache[dir]; loaded {
		return
	}

	ig.cache[dir] = nil

	gitignorePath := filepath.Join(dir, ".gitignore")
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
		ig.cache[dir] = patterns
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
