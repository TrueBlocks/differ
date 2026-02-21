package main

import (
	"os"
	"path/filepath"
	"sort"
)

type fileEntry struct {
	relPath string
	info    os.FileInfo
}

func walkTree(root string, ig *ignorer) ([]fileEntry, error) {
	var entries []fileEntry

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}

		if path == root {
			return nil
		}

		isDir := info.IsDir()
		if ig.isExcluded(path, root, isDir) {
			if isDir {
				return filepath.SkipDir
			}
			return nil
		}

		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}

		entries = append(entries, fileEntry{
			relPath: rel,
			info:    info,
		})

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].relPath < entries[j].relPath
	})

	return entries, nil
}
