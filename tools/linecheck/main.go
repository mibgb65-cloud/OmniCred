package main

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const maximumLines = 400

var checkedExtensions = map[string]bool{
	".go": true, ".ts": true, ".tsx": true, ".js": true,
	".jsx": true, ".css": true, ".sql": true,
}

var skippedDirectories = map[string]bool{
	".git": true, ".cache": true, "node_modules": true,
	"dist": true, "coverage": true,
}

type violation struct {
	path  string
	lines int
}

func main() {
	violations, err := inspect(".")
	if err != nil {
		fmt.Fprintln(os.Stderr, "line check failed:", err)
		os.Exit(2)
	}
	if len(violations) == 0 {
		fmt.Printf("line check passed: all managed code files are within %d lines\n", maximumLines)
		return
	}
	for _, item := range violations {
		fmt.Fprintf(os.Stderr, "%s: %d lines (maximum %d)\n", item.path, item.lines, maximumLines)
	}
	os.Exit(1)
}

func inspect(root string) ([]violation, error) {
	violations := make([]violation, 0)
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			if path != root && skippedDirectories[entry.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if !checkedExtensions[strings.ToLower(filepath.Ext(path))] {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if generated(content) {
			return nil
		}
		lines := bytes.Count(content, []byte{'\n'})
		if len(content) > 0 && content[len(content)-1] != '\n' {
			lines++
		}
		if lines > maximumLines {
			violations = append(violations, violation{path: filepath.ToSlash(path), lines: lines})
		}
		return nil
	})
	sort.Slice(violations, func(i, j int) bool { return violations[i].path < violations[j].path })
	return violations, err
}

func generated(content []byte) bool {
	if len(content) > 2048 {
		content = content[:2048]
	}
	text := string(content)
	return strings.Contains(text, "Code generated") && strings.Contains(text, "DO NOT EDIT")
}
