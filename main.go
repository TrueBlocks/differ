package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	cfg := loadConfig()

	pathA, suffix, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(2)
	}

	pathB, err := computeMirrorPath(pathA, suffix, cfg.PathComponent)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(2)
	}

	if _, err := os.Stat(pathB); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: mirror path does not exist: %s\n", pathB)
		os.Exit(2)
	}

	ignorer := newIgnorer(cfg.AlwaysExclude)

	listA, err := walkTree(pathA, ignorer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking %s: %s\n", pathA, err)
		os.Exit(1)
	}

	listB, err := walkTree(pathB, ignorer)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking %s: %s\n", pathB, err)
		os.Exit(1)
	}

	diffs := computeDiff(listA, listB)
	if len(diffs) == 0 {
		fmt.Println("No differences found.")
		os.Exit(0)
	}

	printDiffs(diffs)
	os.Exit(1)
}

func parseArgs(args []string) (pathA string, suffix int, err error) {
	suffix = 2

	switch len(args) {
	case 0:
		pathA, err = os.Getwd()
		if err != nil {
			return "", 0, fmt.Errorf("cannot get working directory: %w", err)
		}
	case 1:
		pathA, err = filepath.Abs(args[0])
		if err != nil {
			return "", 0, fmt.Errorf("cannot resolve path %q: %w", args[0], err)
		}
	case 2:
		pathA, err = filepath.Abs(args[0])
		if err != nil {
			return "", 0, fmt.Errorf("cannot resolve path %q: %w", args[0], err)
		}
		suffix, err = strconv.Atoi(args[1])
		if err != nil {
			return "", 0, fmt.Errorf("second argument must be a number, got %q", args[1])
		}
		if suffix < 1 {
			return "", 0, fmt.Errorf("suffix must be a positive number, got %d", suffix)
		}
	default:
		return "", 0, fmt.Errorf("usage: differ [path] [number]")
	}

	if _, statErr := os.Stat(pathA); os.IsNotExist(statErr) {
		return "", 0, fmt.Errorf("path does not exist: %s", pathA)
	}

	return pathA, suffix, nil
}

func computeMirrorPath(pathA string, suffix int, component string) (string, error) {
	parts := strings.Split(pathA, string(os.PathSeparator))
	found := false
	for i, p := range parts {
		if p == component {
			parts[i] = fmt.Sprintf("%s.%d", component, suffix)
			found = true
			break
		}
	}

	if !found {
		return "", fmt.Errorf("path %q does not contain a %q component", pathA, component)
	}

	result := strings.Join(parts, string(os.PathSeparator))
	if pathA[0] == '/' {
		result = "/" + strings.TrimLeft(result, string(os.PathSeparator))
	}
	return result, nil
}
