package main

import (
	"path/filepath"
	"testing"
)

var programPath, _ = filepath.Abs("fixtures/main.go")

func Test_findDependencyDirs(t *testing.T) {
	dirs, err := findDependencyDirs(programPath)
	if err != nil {
		t.Fatalf("find deps dirs error: %v", err)
	}
	t.Log(dirs)
}

func Test_findGoModPath(t *testing.T) {
	path, err := findGoModPath(programPath)
	if err != nil {
		t.Fatalf("find go mod path error: %v", err)
	}
	t.Log(path)
}
