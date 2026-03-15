package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	exitOK       = 0
	exitDiff     = 1
	exitUsageErr = 2
)

type options struct {
	useDate   bool
	useHashes bool
	verbose   bool
}

func main() {
	cfg := loadConfig()

	pathA, suffix, opts, err := parseArgs(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(exitUsageErr)
	}

	pathB, err := computeMirrorPath(pathA, suffix)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(exitUsageErr)
	}

	if _, err := os.Stat(pathB); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "Error: mirror path does not exist: %s\n", pathB)
		os.Exit(exitUsageErr)
	}

	ignorer := newIgnorer(cfg.AlwaysExclude, pathA)

	listA, err := walkTree(pathA, ignorer, "A", opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking %s: %s\n", pathA, err)
		os.Exit(exitDiff)
	}

	listB, err := walkTree(pathB, ignorer, "B", opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error walking %s: %s\n", pathB, err)
		os.Exit(exitDiff)
	}

	diffs := computeDiff(listA, listB, pathA, pathB, opts)
	if len(diffs) == 0 {
		fmt.Println("No differences found.")
		os.Exit(exitOK)
	}

	syncDocxNotText(diffs, pathA, pathB)
	syncModes(diffs, pathA, pathB)

	printDiffs(diffs, opts)
	os.Exit(exitDiff)
}

func parseArgs(args []string) (pathA string, suffix int, opts options, err error) {
	suffix = 2

	var positional []string
	for _, arg := range args {
		if arg == "--use-date" {
			opts.useDate = true
		} else if arg == "--hashes" {
			opts.useHashes = true
		} else if arg == "--verbose" {
			opts.verbose = true
		} else {
			positional = append(positional, arg)
		}
	}
	args = positional

	switch len(args) {
	case 0:
		pathA, err = os.Getwd()
		if err != nil {
			return "", 0, opts, fmt.Errorf("cannot get working directory: %w", err)
		}
	case 1:
		pathA, err = filepath.Abs(args[0])
		if err != nil {
			return "", 0, opts, fmt.Errorf("cannot resolve path %q: %w", args[0], err)
		}
	case 2:
		pathA, err = filepath.Abs(args[0])
		if err != nil {
			return "", 0, opts, fmt.Errorf("cannot resolve path %q: %w", args[0], err)
		}
		suffix, err = strconv.Atoi(args[1])
		if err != nil {
			return "", 0, opts, fmt.Errorf("second argument must be a number, got %q", args[1])
		}
		if suffix < 1 {
			return "", 0, opts, fmt.Errorf("suffix must be a positive number, got %d", suffix)
		}
	default:
		return "", 0, opts, fmt.Errorf("usage: differ [path] [number]")
	}

	if _, statErr := os.Stat(pathA); os.IsNotExist(statErr) {
		return "", 0, opts, fmt.Errorf("path does not exist: %s", pathA)
	}

	return pathA, suffix, opts, nil
}

func computeMirrorPath(pathA string, suffix int) (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot determine home directory: %w", err)
	}

	if !strings.HasPrefix(pathA, home+string(os.PathSeparator)) {
		return "", fmt.Errorf("path %q is not under home directory %q", pathA, home)
	}

	rel := strings.TrimPrefix(pathA, home+string(os.PathSeparator))
	parts := strings.SplitN(rel, string(os.PathSeparator), 2)
	component := parts[0]
	if component == "" {
		return "", fmt.Errorf("path %q has no directory component after home", pathA)
	}

	newComponent := fmt.Sprintf("%s.%d", component, suffix)
	result := filepath.Join(home, newComponent)
	if len(parts) > 1 {
		result = filepath.Join(result, parts[1])
	}
	return result, nil
}
