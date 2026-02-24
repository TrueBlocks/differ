package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"strings"
)

type docxCategory int

const (
	catText docxCategory = iota
	catMeta
	catStyle
	catMedia
	catOther
)

func categorizeDocxFile(name string) docxCategory {
	lower := strings.ToLower(name)
	switch {
	case lower == "word/document.xml",
		lower == "word/footnotes.xml",
		lower == "word/endnotes.xml",
		lower == "word/comments.xml",
		strings.HasPrefix(lower, "word/header"),
		strings.HasPrefix(lower, "word/footer"):
		return catText
	case strings.HasPrefix(lower, "docprops/"),
		lower == "[content_types].xml",
		lower == "_rels/.rels",
		strings.HasPrefix(lower, "word/_rels/"):
		return catMeta
	case lower == "word/styles.xml",
		lower == "word/settings.xml",
		lower == "word/fonttable.xml",
		lower == "word/numbering.xml",
		lower == "word/websettings.xml",
		strings.HasPrefix(lower, "word/theme/"):
		return catStyle
	case strings.HasPrefix(lower, "word/media/"):
		return catMedia
	default:
		return catOther
	}
}

func categoryLabel(c docxCategory) string {
	switch c {
	case catText:
		return "text"
	case catMeta:
		return "meta"
	case catStyle:
		return "style"
	case catMedia:
		return "media"
	default:
		return "other"
	}
}

func hashZipEntry(f *zip.File) (string, error) {
	rc, err := f.Open()
	if err != nil {
		return "", err
	}
	defer rc.Close()
	h := sha256.New()
	if _, err := io.Copy(h, rc); err != nil {
		return "", err
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func readZipEntry(r *zip.ReadCloser, name string) ([]byte, error) {
	for _, f := range r.File {
		if f.Name == name {
			rc, err := f.Open()
			if err != nil {
				return nil, err
			}
			defer rc.Close()
			return io.ReadAll(rc)
		}
	}
	return nil, fmt.Errorf("not found: %s", name)
}

func extractPlainText(xmlData []byte) string {
	var buf strings.Builder
	dec := xml.NewDecoder(strings.NewReader(string(xmlData)))
	inText := false
	for {
		tok, err := dec.Token()
		if err != nil {
			break
		}
		switch t := tok.(type) {
		case xml.StartElement:
			if t.Name.Local == "t" {
				inText = true
			}
		case xml.EndElement:
			if t.Name.Local == "t" {
				inText = false
			}
		case xml.CharData:
			if inText {
				buf.Write(t)
			}
		}
	}
	return buf.String()
}

type docxFileDiff struct {
	name     string
	category docxCategory
	reason   string
	textDiff []string
}

type docxResult struct {
	label   string
	details []docxFileDiff
}

func analyzeDocx(pathA, pathB string) docxResult {
	rA, errA := zip.OpenReader(pathA)
	rB, errB := zip.OpenReader(pathB)
	if errA != nil || errB != nil {
		return docxResult{label: "docx:err"}
	}
	defer rA.Close()
	defer rB.Close()

	type entryInfo struct {
		size uint64
		hash string
	}

	buildMap := func(r *zip.ReadCloser) map[string]entryInfo {
		m := make(map[string]entryInfo)
		for _, f := range r.File {
			h, _ := hashZipEntry(f)
			m[f.Name] = entryInfo{size: f.UncompressedSize64, hash: h}
		}
		return m
	}

	mapA := buildMap(rA)
	mapB := buildMap(rB)

	diffCats := make(map[docxCategory]bool)
	var fileDiffs []docxFileDiff

	textContentFiles := map[string]bool{
		"word/document.xml":  true,
		"word/footnotes.xml": true,
		"word/endnotes.xml":  true,
		"word/comments.xml":  true,
	}
	for i := 1; i <= 9; i++ {
		textContentFiles[fmt.Sprintf("word/header%d.xml", i)] = true
		textContentFiles[fmt.Sprintf("word/footer%d.xml", i)] = true
	}

	for name, a := range mapA {
		b, ok := mapB[name]
		if !ok {
			cat := categorizeDocxFile(name)
			if cat == catText && textContentFiles[strings.ToLower(name)] {
				dataA, errA := readZipEntry(rA, name)
				if errA == nil {
					textA := extractPlainText(dataA)
					if textA == "" {
						cat = catMeta
						fileDiffs = append(fileDiffs, docxFileDiff{name: name, category: cat, reason: "only in A (empty)"})
					} else {
						diff := computeTextDiff(textA, "", name)
						fileDiffs = append(fileDiffs, docxFileDiff{name: name, category: cat, reason: "only in A", textDiff: diff})
					}
				} else {
					fileDiffs = append(fileDiffs, docxFileDiff{name: name, category: cat, reason: "only in A"})
				}
			} else {
				fileDiffs = append(fileDiffs, docxFileDiff{name: name, category: cat, reason: "only in A"})
			}
			diffCats[cat] = true
			continue
		}
		if a.hash != b.hash {
			cat := categorizeDocxFile(name)
			if cat == catText && textContentFiles[strings.ToLower(name)] {
				dataA, errA := readZipEntry(rA, name)
				dataB, errB := readZipEntry(rB, name)
				if errA == nil && errB == nil {
					textA := extractPlainText(dataA)
					textB := extractPlainText(dataB)
					if textA == textB {
						cat = catStyle
						fileDiffs = append(fileDiffs, docxFileDiff{name: name, category: cat, reason: "markup only (text identical)"})
					} else {
						diff := computeTextDiff(textA, textB, name)
						fileDiffs = append(fileDiffs, docxFileDiff{name: name, category: cat, reason: "text content differs", textDiff: diff})
					}
				} else {
					fileDiffs = append(fileDiffs, docxFileDiff{name: name, category: cat, reason: "hash differs"})
				}
			} else {
				fileDiffs = append(fileDiffs, docxFileDiff{name: name, category: cat, reason: "hash differs"})
			}
			diffCats[cat] = true
		}
	}
	for name := range mapB {
		if _, ok := mapA[name]; !ok {
			cat := categorizeDocxFile(name)
			if cat == catText && textContentFiles[strings.ToLower(name)] {
				dataB, errB := readZipEntry(rB, name)
				if errB == nil {
					textB := extractPlainText(dataB)
					if textB == "" {
						cat = catMeta
						fileDiffs = append(fileDiffs, docxFileDiff{name: name, category: cat, reason: "only in B (empty)"})
					} else {
						diff := computeTextDiff("", textB, name)
						fileDiffs = append(fileDiffs, docxFileDiff{name: name, category: cat, reason: "only in B", textDiff: diff})
					}
				} else {
					fileDiffs = append(fileDiffs, docxFileDiff{name: name, category: cat, reason: "only in B"})
				}
			} else {
				fileDiffs = append(fileDiffs, docxFileDiff{name: name, category: cat, reason: "only in B"})
			}
			diffCats[cat] = true
		}
	}

	if len(diffCats) == 0 {
		return docxResult{label: "docx:identical", details: fileDiffs}
	}

	if diffCats[catText] {
		return docxResult{label: "docx:text", details: fileDiffs}
	}
	return docxResult{label: "docx:not-text", details: fileDiffs}
}

func isDocx(relPath string) bool {
	return strings.ToLower(filepath.Ext(relPath)) == ".docx"
}

func computeTextDiff(textA, textB, name string) []string {
	linesA := strings.Split(textA, "\n")
	linesB := strings.Split(textB, "\n")

	if len(linesA) == 1 && len(linesB) == 1 {
		linesA = splitIntoChunks(textA, 80)
		linesB = splitIntoChunks(textB, 80)
	}

	var result []string
	matches := lcs(linesA, linesB)

	idxA, idxB := 0, 0
	for _, m := range matches {
		for idxA < m.a {
			result = append(result, fmt.Sprintf("  - %s", linesA[idxA]))
			idxA++
		}
		for idxB < m.b {
			result = append(result, fmt.Sprintf("  + %s", linesB[idxB]))
			idxB++
		}
		idxA++
		idxB++
	}
	for idxA < len(linesA) {
		result = append(result, fmt.Sprintf("  - %s", linesA[idxA]))
		idxA++
	}
	for idxB < len(linesB) {
		result = append(result, fmt.Sprintf("  + %s", linesB[idxB]))
		idxB++
	}

	return result
}

type match struct {
	a, b int
}

func lcs(a, b []string) []match {
	n, m := len(a), len(b)
	if n == 0 || m == 0 {
		return nil
	}

	dp := make([][]int, n+1)
	for i := range dp {
		dp[i] = make([]int, m+1)
	}
	for i := 1; i <= n; i++ {
		for j := 1; j <= m; j++ {
			if a[i-1] == b[j-1] {
				dp[i][j] = dp[i-1][j-1] + 1
			} else if dp[i-1][j] >= dp[i][j-1] {
				dp[i][j] = dp[i-1][j]
			} else {
				dp[i][j] = dp[i][j-1]
			}
		}
	}

	var result []match
	i, j := n, m
	for i > 0 && j > 0 {
		if a[i-1] == b[j-1] {
			result = append(result, match{i - 1, j - 1})
			i--
			j--
		} else if dp[i-1][j] >= dp[i][j-1] {
			i--
		} else {
			j--
		}
	}

	for l, r := 0, len(result)-1; l < r; l, r = l+1, r-1 {
		result[l], result[r] = result[r], result[l]
	}
	return result
}

func splitIntoChunks(s string, size int) []string {
	var chunks []string
	for len(s) > size {
		chunks = append(chunks, s[:size])
		s = s[size:]
	}
	if len(s) > 0 {
		chunks = append(chunks, s)
	}
	return chunks
}
