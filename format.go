package main

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"

	"golang.org/x/term"
)

const (
	minTermWidth     = 40
	defaultTermWidth = 80
	minTruncLen      = 3
	minDetailCol     = 6
	maxDetailCol     = 24
	groupColWidth    = 18
	fixedColOverhead = 29
	minFileCol       = 10
)

func termWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w < minTermWidth {
		return defaultTermWidth
	}
	return w
}

func truncatePath(path string, maxLen int) string {
	if len(path) <= maxLen {
		return path
	}
	if maxLen <= minTruncLen {
		return path[:maxLen]
	}
	return "..." + path[len(path)-(maxLen-minTruncLen):]
}

func formatEntry(e fileEntry, opts options) string {
	info := e.info
	mode := info.Mode().String()
	size := info.Size()

	typeIndicator := ""
	if info.IsDir() {
		typeIndicator = "/"
	} else if info.Mode()&os.ModeSymlink != 0 {
		typeIndicator = "@"
	}

	if opts.useDate {
		mod := info.ModTime().Format("Jan _2 15:04")
		return fmt.Sprintf("%s %8d %s %s%s", mode, size, mod, e.relPath, typeIndicator)
	}
	return fmt.Sprintf("%s %8d %s%s", mode, size, e.relPath, typeIndicator)
}

func colorize(code, s string) string {
	if !isTTY() {
		return s
	}
	return "\033[" + code + "m" + s + "\033[0m"
}

func colorRed(s string) string    { return colorize("31", s) }
func colorGreen(s string) string  { return colorize("32", s) }
func colorYellow(s string) string { return colorize("33", s) }
func colorCyan(s string) string   { return colorize("36", s) }

var (
	ttyOnce sync.Once
	ttyVal  bool
)

func isTTY() bool {
	ttyOnce.Do(func() {
		fi, err := os.Stdout.Stat()
		if err != nil {
			return
		}
		ttyVal = (fi.Mode() & os.ModeCharDevice) != 0
	})
	return ttyVal
}

func splitGroupAndFile(relPath string) (group string, file string) {
	parts := strings.SplitN(relPath, string(os.PathSeparator), 2)
	if len(parts) == 1 {
		return ".", parts[0]
	}
	return parts[0], parts[1]
}

func sortDiffEntries(entries []diffEntry) {
	sort.Slice(entries, func(i, j int) bool {
		di := strings.Join(shortDetails(entries[i].details), " ")
		dj := strings.Join(shortDetails(entries[j].details), " ")
		if di != dj {
			return di < dj
		}
		gi, fi := splitGroupAndFile(entries[i].relPath)
		gj, fj := splitGroupAndFile(entries[j].relPath)
		if gi != gj {
			return gi < gj
		}
		return fi < fj
	})
}

func printDiffs(diffs []diffEntry, opts options) {
	var onlyA, onlyB, changed, synced []diffEntry
	for _, d := range diffs {
		switch d.kind {
		case diffOnlyA:
			onlyA = append(onlyA, d)
		case diffOnlyB:
			onlyB = append(onlyB, d)
		case diffChanged:
			changed = append(changed, d)
		case diffSynced:
			synced = append(synced, d)
		}
	}

	sortDiffEntries(onlyA)
	sortDiffEntries(onlyB)
	sortDiffEntries(changed)

	w := termWidth()

	allDiffs := append(append(onlyA, onlyB...), changed...)
	detailMax := minDetailCol
	for _, d := range allDiffs {
		s := detailString(d.details)
		if len(s) > detailMax {
			detailMax = len(s)
		}
	}
	if detailMax > maxDetailCol {
		detailMax = maxDetailCol
	}

	groupMax := groupColWidth
	// fixed: 2(pad) + 1(S) + 2 + detail + 2 + group + 2 + 8(sizeA) + 2 + 8(sizeB) + 2 = 47 + detail + group
	fileMax := w - fixedColOverhead - detailMax - groupMax
	if fileMax < minFileCol {
		fileMax = minFileCol
	}

	header := fmt.Sprintf("  %s  %-*s  %-*s  %-*s  %8s  %8s",
		"S", detailMax, "DETAIL", groupMax, "GROUP", fileMax, "FILE", "SIZE_A", "SIZE_B")
	sep := strings.Repeat("-", w)

	if len(onlyA) > 0 {
		fmt.Println(colorCyan("=== Only in A ==="))
		fmt.Println(colorCyan(header))
		fmt.Println(colorCyan(sep))
		for _, d := range onlyA {
			path := d.entryA.relPath
			if d.entryA.info.IsDir() {
				path += "/"
			}
			group, file := splitGroupAndFile(path)
			fmt.Println(colorRed(fmt.Sprintf("  %s  %-*s  %-*s  %-*s  %8d  %8s",
				"-", detailMax, "", groupMax, truncatePath(group, groupMax), fileMax, truncatePath(file, fileMax), d.entryA.info.Size(), "-")))
		}
		fmt.Println()
	}

	if len(onlyB) > 0 {
		fmt.Println(colorCyan("=== Only in B ==="))
		fmt.Println(colorCyan(header))
		fmt.Println(colorCyan(sep))
		for _, d := range onlyB {
			path := d.entryB.relPath
			if d.entryB.info.IsDir() {
				path += "/"
			}
			group, file := splitGroupAndFile(path)
			fmt.Println(colorGreen(fmt.Sprintf("  %s  %-*s  %-*s  %-*s  %8s  %8d",
				"+", detailMax, "", groupMax, truncatePath(group, groupMax), fileMax, truncatePath(file, fileMax), "-", d.entryB.info.Size())))
		}
		fmt.Println()
	}

	if len(changed) > 0 {
		fmt.Println(colorCyan("=== Changed ==="))
		fmt.Println(colorCyan(header))
		fmt.Println(colorCyan(sep))
		for _, d := range changed {
			path := d.relPath
			if d.entryA.info.IsDir() {
				path += "/"
			}
			group, file := splitGroupAndFile(path)
			detail := detailString(d.details)
			if len(detail) > detailMax {
				detail = firstDetail(d.details)
			}
			fmt.Println(colorYellow(fmt.Sprintf("  %s  %-*s  %-*s  %-*s  %8d  %8d",
				"~", detailMax, detail, groupMax, truncatePath(group, groupMax), fileMax, truncatePath(file, fileMax), d.entryA.info.Size(), d.entryB.info.Size())))
			if opts.verbose && len(d.docxDetails) > 0 {
				for _, dd := range d.docxDetails {
					fmt.Println(colorYellow(fmt.Sprintf("      %-8s  %s  (%s)",
						categoryLabel(dd.category), dd.name, dd.reason)))
					if len(dd.textDiff) > 0 {
						for _, line := range dd.textDiff {
							if strings.HasPrefix(line, "  -") {
								fmt.Println(colorRed(line))
							} else if strings.HasPrefix(line, "  +") {
								fmt.Println(colorGreen(line))
							} else {
								fmt.Println(line)
							}
						}
					}
				}
			}
		}
		fmt.Println()
	}

	fmt.Printf("Summary: %d only in A, %d only in B, %d changed",
		len(onlyA), len(onlyB), len(changed))
	if len(synced) > 0 {
		fmt.Printf(", %d synced", len(synced))
	}
	fmt.Println()
}

func shortDetails(details []string) []string {
	hasDocx := false
	for _, d := range details {
		if strings.HasPrefix(d, "docx:") {
			hasDocx = true
			break
		}
	}

	var short []string
	for _, d := range details {
		if strings.HasPrefix(d, "size:") {
			if hasDocx {
				continue
			}
			short = append(short, "size")
		} else if strings.HasPrefix(d, "mode:") {
			short = append(short, "mode")
		} else if strings.HasPrefix(d, "hash:") {
			short = append(short, "hash")
		} else if strings.HasPrefix(d, "modified:") {
			short = append(short, "date")
		} else if strings.HasPrefix(d, "docx:") {
			short = append(short, d)
		} else {
			short = append(short, d)
		}
	}
	return short
}

func detailString(details []string) string {
	return strings.Join(shortDetails(details), " ")
}

func firstDetail(details []string) string {
	short := shortDetails(details)
	if len(short) == 0 {
		return ""
	}
	return short[0]
}
