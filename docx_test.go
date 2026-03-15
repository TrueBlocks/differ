package main

import (
	"testing"
)

func TestSplitIntoChunks(t *testing.T) {
	tests := []struct {
		name string
		s    string
		size int
		want []string
	}{
		{"exact multiple", "aabbcc", 2, []string{"aa", "bb", "cc"}},
		{"remainder", "aabbc", 2, []string{"aa", "bb", "c"}},
		{"larger than input", "abc", 10, []string{"abc"}},
		{"empty string", "", 5, nil},
		{"size 1", "abc", 1, []string{"a", "b", "c"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitIntoChunks(tt.s, tt.size)
			if len(got) != len(tt.want) {
				t.Fatalf("splitIntoChunks(%q, %d) = %v, want %v", tt.s, tt.size, got, tt.want)
			}
			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("splitIntoChunks(%q, %d)[%d] = %q, want %q", tt.s, tt.size, i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestLCS(t *testing.T) {
	tests := []struct {
		name string
		a, b []string
		want int // number of matches
	}{
		{"identical", []string{"a", "b", "c"}, []string{"a", "b", "c"}, 3},
		{"no common", []string{"a", "b"}, []string{"c", "d"}, 0},
		{"partial overlap", []string{"a", "b", "c", "d"}, []string{"b", "d"}, 2},
		{"empty a", []string{}, []string{"a"}, 0},
		{"empty b", []string{"a"}, []string{}, 0},
		{"both empty", []string{}, []string{}, 0},
		{"single match", []string{"x", "a", "y"}, []string{"p", "a", "q"}, 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := lcs(tt.a, tt.b)
			if len(got) != tt.want {
				t.Errorf("lcs(%v, %v) returned %d matches, want %d", tt.a, tt.b, len(got), tt.want)
			}
		})
	}
}

func TestCategorizeDocxFile(t *testing.T) {
	tests := []struct {
		name string
		file string
		want docxCategory
	}{
		{"document", "word/document.xml", catText},
		{"footnotes", "word/footnotes.xml", catText},
		{"header", "word/header1.xml", catText},
		{"footer", "word/footer2.xml", catText},
		{"styles", "word/styles.xml", catStyle},
		{"settings", "word/settings.xml", catStyle},
		{"theme", "word/theme/theme1.xml", catStyle},
		{"media", "word/media/image1.png", catMedia},
		{"rels", "_rels/.rels", catMeta},
		{"content types", "[Content_Types].xml", catMeta},
		{"docprops", "docProps/core.xml", catMeta},
		{"unknown", "customXml/item1.xml", catOther},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := categorizeDocxFile(tt.file)
			if got != tt.want {
				t.Errorf("categorizeDocxFile(%q) = %d, want %d", tt.file, got, tt.want)
			}
		})
	}
}

func TestExtractPlainText(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body>
    <w:p><w:r><w:t>Hello </w:t></w:r><w:r><w:t>world</w:t></w:r></w:p>
  </w:body>
</w:document>`

	got := extractPlainText([]byte(xml))
	if got != "Hello world" {
		t.Errorf("extractPlainText returned %q, want %q", got, "Hello world")
	}
}

func TestExtractPlainText_Empty(t *testing.T) {
	xml := `<?xml version="1.0" encoding="UTF-8"?>
<w:document xmlns:w="http://schemas.openxmlformats.org/wordprocessingml/2006/main">
  <w:body></w:body>
</w:document>`

	got := extractPlainText([]byte(xml))
	if got != "" {
		t.Errorf("extractPlainText returned %q, want empty", got)
	}
}

func TestIsDocx(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"essay.docx", true},
		{"essay.DOCX", true},
		{"essay.txt", false},
		{"essay.doc", false},
		{"path/to/file.docx", true},
	}
	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			if got := isDocx(tt.path); got != tt.want {
				t.Errorf("isDocx(%q) = %v, want %v", tt.path, got, tt.want)
			}
		})
	}
}
