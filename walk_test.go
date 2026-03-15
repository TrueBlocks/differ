package main

import (
	"os"
	"path/filepath"
	"testing"
)

func TestHashFile_Success(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	h, err := hashFile(path)
	if err != nil {
		t.Fatalf("hashFile(%q) returned error: %v", path, err)
	}
	if len(h) != 64 { // SHA256 hex length
		t.Errorf("hash length = %d, want 64", len(h))
	}
}

func TestHashFile_Deterministic(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test.txt")
	if err := os.WriteFile(path, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	h1, _ := hashFile(path)
	h2, _ := hashFile(path)
	if h1 != h2 {
		t.Errorf("hashFile not deterministic: %q != %q", h1, h2)
	}
}

func TestHashFile_DifferentContent(t *testing.T) {
	dir := t.TempDir()
	path1 := filepath.Join(dir, "a.txt")
	path2 := filepath.Join(dir, "b.txt")
	os.WriteFile(path1, []byte("hello"), 0644)
	os.WriteFile(path2, []byte("world"), 0644)

	h1, _ := hashFile(path1)
	h2, _ := hashFile(path2)
	if h1 == h2 {
		t.Error("different content produced same hash")
	}
}

func TestHashFile_NotFound(t *testing.T) {
	_, err := hashFile("/nonexistent/path/file.txt")
	if err == nil {
		t.Error("hashFile on nonexistent file should return error")
	}
}
