package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
)

type fileEntry struct {
	relPath string
	info    os.FileInfo
	hash    string
}

func walkTree(root string, ig *ignorer, label string) ([]fileEntry, error) {
	var entries []fileEntry
	count := 0

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if path == root {
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		isDir := info.IsDir()
		if ig.isExcluded(rel, isDir) {
			if isDir {
				return filepath.SkipDir
			}
			return nil
		}

		count++
		entry := fileEntry{
			relPath: rel,
			info:    info,
		}
		if useHashes && !isDir {
			fmt.Fprintf(os.Stderr, "\r  Scanning %s: %d files (hashing)...", label, count)
			entry.hash = hashFile(path)
		} else {
			fmt.Fprintf(os.Stderr, "\r  Scanning %s: %d files...", label, count)
		}
		entries = append(entries, entry)

		return nil
	})

	if err != nil {
		return nil, err
	}

	if count > 0 {
		fmt.Fprintf(os.Stderr, "\r  Scanning %s: %d files... done.                \n", label, count)
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].relPath < entries[j].relPath
	})

	return entries, nil
}

func hashFile(path string) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return hex.EncodeToString(h.Sum(nil))
}
