package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

const maxDepth = 1

func main() {
	if len(os.Args) < 2 {
		return
	}
	srcDir := filepath.ToSlash(os.Args[1])
	srcDir = filepath.Clean(srcDir)
	root := strings.TrimRight(srcDir, "/\\")
	rootDepth := strings.Count(root, "/")

	fmt.Println("root dir: ", root)
	fmt.Println("root depth: ", rootDepth)

	pathDepth := func(p string) int {
		return strings.Count(p, "/") - rootDepth
	}

	err := filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return fmt.Errorf("error at %s: %v", path, err)
		}
		path = filepath.ToSlash(path) // path normalization (unix)
		depth := pathDepth(path)

		// skip root folder
		if path == root {
			return nil
		}

		// directories
		if d.IsDir() {
			if d.Name() == ".git" {
				return filepath.SkipDir
			}
			if depth >= maxDepth {
				info, _ := d.Info()
				fmt.Printf("%s %d %v %v %v\n", info.Name(), info.Size(), info.Mode(), info.ModTime(), info.IsDir())

				return filepath.SkipDir
			}
			return nil
		}

		fmt.Printf("%d %s\n", depth, path)
		return nil
	})
	if err != nil {
		fmt.Println("walk error:", err)
	}
}
