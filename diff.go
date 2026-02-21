package main

import (
	"fmt"
)

type diffKind int

const (
	diffOnlyA diffKind = iota
	diffOnlyB
	diffChanged
)

type diffEntry struct {
	kind    diffKind
	relPath string
	entryA  *fileEntry
	entryB  *fileEntry
	details []string
}

func computeDiff(listA, listB []fileEntry) []diffEntry {
	mapA := make(map[string]*fileEntry, len(listA))
	for i := range listA {
		mapA[listA[i].relPath] = &listA[i]
	}

	mapB := make(map[string]*fileEntry, len(listB))
	for i := range listB {
		mapB[listB[i].relPath] = &listB[i]
	}

	seen := make(map[string]bool)
	var diffs []diffEntry

	for _, a := range listA {
		seen[a.relPath] = true
		b, inB := mapB[a.relPath]
		if !inB {
			diffs = append(diffs, diffEntry{
				kind:    diffOnlyA,
				relPath: a.relPath,
				entryA:  mapA[a.relPath],
			})
			continue
		}

		changes := compareEntries(mapA[a.relPath], b)
		if len(changes) > 0 {
			diffs = append(diffs, diffEntry{
				kind:    diffChanged,
				relPath: a.relPath,
				entryA:  mapA[a.relPath],
				entryB:  b,
				details: changes,
			})
		}
	}

	for _, b := range listB {
		if !seen[b.relPath] {
			diffs = append(diffs, diffEntry{
				kind:    diffOnlyB,
				relPath: b.relPath,
				entryB:  mapB[b.relPath],
			})
		}
	}

	return diffs
}

func compareEntries(a, b *fileEntry) []string {
	var changes []string

	if a.info.Mode() != b.info.Mode() {
		changes = append(changes, fmt.Sprintf("mode: %s vs %s", a.info.Mode(), b.info.Mode()))
	}

	if !a.info.IsDir() && !b.info.IsDir() {
		if a.info.Size() != b.info.Size() {
			changes = append(changes, fmt.Sprintf("size: %d vs %d", a.info.Size(), b.info.Size()))
		}
	}

	if !a.info.ModTime().Equal(b.info.ModTime()) {
		changes = append(changes, fmt.Sprintf("modified: %s vs %s",
			a.info.ModTime().Format("2006-01-02 15:04:05"),
			b.info.ModTime().Format("2006-01-02 15:04:05")))
	}

	return changes
}
