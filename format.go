package main

import (
	"fmt"
	"os"
	"strings"
)

func formatEntry(e fileEntry) string {
	info := e.info
	mode := info.Mode().String()
	size := info.Size()
	mod := info.ModTime().Format("Jan _2 15:04")

	typeIndicator := ""
	if info.IsDir() {
		typeIndicator = "/"
	} else if info.Mode()&os.ModeSymlink != 0 {
		typeIndicator = "@"
	}

	return fmt.Sprintf("%s %8d %s %s%s", mode, size, mod, e.relPath, typeIndicator)
}

func colorRed(s string) string {
	if !isTTY() {
		return s
	}
	return "\033[31m" + s + "\033[0m"
}

func colorGreen(s string) string {
	if !isTTY() {
		return s
	}
	return "\033[32m" + s + "\033[0m"
}

func colorYellow(s string) string {
	if !isTTY() {
		return s
	}
	return "\033[33m" + s + "\033[0m"
}

func colorCyan(s string) string {
	if !isTTY() {
		return s
	}
	return "\033[36m" + s + "\033[0m"
}

var ttyChecked bool
var ttyResult bool

func isTTY() bool {
	if ttyChecked {
		return ttyResult
	}
	ttyChecked = true
	fi, err := os.Stdout.Stat()
	if err != nil {
		ttyResult = false
		return false
	}
	ttyResult = (fi.Mode() & os.ModeCharDevice) != 0
	return ttyResult
}

func printDiffs(diffs []diffEntry) {
	var onlyA, onlyB, changed []diffEntry
	for _, d := range diffs {
		switch d.kind {
		case diffOnlyA:
			onlyA = append(onlyA, d)
		case diffOnlyB:
			onlyB = append(onlyB, d)
		case diffChanged:
			changed = append(changed, d)
		}
	}

	if len(onlyA) > 0 {
		fmt.Println(colorCyan("=== Only in path A ==="))
		for _, d := range onlyA {
			fmt.Println(colorRed("- " + formatEntry(*d.entryA)))
		}
		fmt.Println()
	}

	if len(onlyB) > 0 {
		fmt.Println(colorCyan("=== Only in path B ==="))
		for _, d := range onlyB {
			fmt.Println(colorGreen("+ " + formatEntry(*d.entryB)))
		}
		fmt.Println()
	}

	if len(changed) > 0 {
		fmt.Println(colorCyan("=== Changed ==="))
		for _, d := range changed {
			fmt.Println(colorYellow("~ " + d.relPath))
			fmt.Println(colorRed("  A: " + formatEntry(*d.entryA)))
			fmt.Println(colorGreen("  B: " + formatEntry(*d.entryB)))
			if len(d.details) > 0 {
				fmt.Println("    " + strings.Join(d.details, ", "))
			}
		}
		fmt.Println()
	}

	fmt.Printf("Summary: %d only in A, %d only in B, %d changed\n",
		len(onlyA), len(onlyB), len(changed))
}
