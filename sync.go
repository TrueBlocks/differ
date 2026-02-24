package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func syncDocxNotText(diffs []diffEntry, rootA, rootB string) {
	var synced int
	for i, d := range diffs {
		if d.kind != diffChanged {
			continue
		}
		detail := detailString(d.details)
		if detail != "docx:not-text" && detail != "docx:identical" {
			continue
		}

		srcPath := filepath.Join(rootA, d.relPath)
		dstPath := filepath.Join(rootB, d.relPath)

		if err := copyFile(srcPath, dstPath); err != nil {
			fmt.Fprintf(os.Stderr, "  sync error: %s: %s\n", d.relPath, err)
			continue
		}

		fmt.Fprintf(os.Stderr, "  synced: %s\n", d.relPath)
		synced++
		diffs[i].kind = diffSynced
	}
	if synced > 0 {
		fmt.Fprintf(os.Stderr, "  %d files synced (A → B)\n\n", synced)
	}
}

func syncModes(diffs []diffEntry, rootA, rootB string) {
	var fixed int
	for i, d := range diffs {
		if d.kind != diffChanged {
			continue
		}
		if d.entryA == nil {
			continue
		}

		hasMode := false
		for _, det := range d.details {
			if strings.HasPrefix(det, "mode:") {
				hasMode = true
				break
			}
		}
		if !hasMode {
			continue
		}

		dstPath := filepath.Join(rootB, d.relPath)
		modeA := d.entryA.info.Mode()
		if err := os.Chmod(dstPath, modeA); err != nil {
			fmt.Fprintf(os.Stderr, "  chmod error: %s: %s\n", d.relPath, err)
			continue
		}

		fmt.Fprintf(os.Stderr, "  chmod: %s → %s\n", d.relPath, modeA)
		fixed++

		var remaining []string
		for _, det := range diffs[i].details {
			if !strings.HasPrefix(det, "mode:") {
				remaining = append(remaining, det)
			}
		}
		if len(remaining) == 0 {
			diffs[i].kind = diffSynced
		} else {
			diffs[i].details = remaining
		}
	}
	if fixed > 0 {
		fmt.Fprintf(os.Stderr, "  %d modes fixed (A → B)\n\n", fixed)
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	info, err := in.Stat()
	if err != nil {
		return err
	}

	out, err := os.OpenFile(dst, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, info.Mode())
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}

	return out.Close()
}
