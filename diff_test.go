package main

import (
	"testing"
)

func TestTruncHash(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{"full sha256", "abcdef123456789abcdef", "abcdef123456"},
		{"exactly 12", "abcdef123456", "abcdef123456"},
		{"short hash", "abc", "abc"},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := truncHash(tt.in)
			if got != tt.want {
				t.Errorf("truncHash(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestComputeDiff_OnlyA(t *testing.T) {
	listA := []fileEntry{{relPath: "a.txt", info: fakeInfo{name: "a.txt", size: 10}}}
	listB := []fileEntry{}
	opts := options{}

	diffs := computeDiff(listA, listB, "/tmp/a", "/tmp/b", opts)

	if len(diffs) != 1 {
		t.Fatalf("got %d diffs, want 1", len(diffs))
	}
	if diffs[0].kind != diffOnlyA {
		t.Errorf("kind = %d, want diffOnlyA (%d)", diffs[0].kind, diffOnlyA)
	}
	if diffs[0].relPath != "a.txt" {
		t.Errorf("relPath = %q, want %q", diffs[0].relPath, "a.txt")
	}
}

func TestComputeDiff_OnlyB(t *testing.T) {
	listA := []fileEntry{}
	listB := []fileEntry{{relPath: "b.txt", info: fakeInfo{name: "b.txt", size: 20}}}
	opts := options{}

	diffs := computeDiff(listA, listB, "/tmp/a", "/tmp/b", opts)

	if len(diffs) != 1 {
		t.Fatalf("got %d diffs, want 1", len(diffs))
	}
	if diffs[0].kind != diffOnlyB {
		t.Errorf("kind = %d, want diffOnlyB (%d)", diffs[0].kind, diffOnlyB)
	}
}

func TestComputeDiff_SizeDiffers(t *testing.T) {
	listA := []fileEntry{{relPath: "f.txt", info: fakeInfo{name: "f.txt", size: 10}}}
	listB := []fileEntry{{relPath: "f.txt", info: fakeInfo{name: "f.txt", size: 20}}}
	opts := options{}

	diffs := computeDiff(listA, listB, "/tmp/a", "/tmp/b", opts)

	if len(diffs) != 1 {
		t.Fatalf("got %d diffs, want 1", len(diffs))
	}
	if diffs[0].kind != diffChanged {
		t.Errorf("kind = %d, want diffChanged (%d)", diffs[0].kind, diffChanged)
	}
}

func TestComputeDiff_Identical(t *testing.T) {
	listA := []fileEntry{{relPath: "f.txt", info: fakeInfo{name: "f.txt", size: 10}}}
	listB := []fileEntry{{relPath: "f.txt", info: fakeInfo{name: "f.txt", size: 10}}}
	opts := options{}

	diffs := computeDiff(listA, listB, "/tmp/a", "/tmp/b", opts)

	if len(diffs) != 0 {
		t.Errorf("got %d diffs, want 0", len(diffs))
	}
}
