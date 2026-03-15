package main

import (
	"os"
	"time"
)

// fakeInfo implements os.FileInfo for testing without touching the filesystem.
type fakeInfo struct {
	name string
	size int64
	mode os.FileMode
	mod  time.Time
	dir  bool
}

func (f fakeInfo) Name() string      { return f.name }
func (f fakeInfo) Size() int64       { return f.size }
func (f fakeInfo) Mode() os.FileMode { return f.mode }
func (f fakeInfo) ModTime() time.Time {
	if f.mod.IsZero() {
		return time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	return f.mod
}
func (f fakeInfo) IsDir() bool      { return f.dir }
func (f fakeInfo) Sys() interface{} { return nil }
