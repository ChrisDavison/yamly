package walk

import (
	"os"
	"path/filepath"
)

// Walk returns all *.md file paths found recursively under root,
// skipping any directories whose name matches an entry in excludes.
func Walk(root string, excludes []string) ([]string, error) {
	excluded := make(map[string]bool, len(excludes))
	for _, e := range excludes {
		excluded[e] = true
	}

	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && excluded[filepath.Base(path)] {
			return filepath.SkipDir
		}
		if !d.IsDir() && filepath.Ext(path) == ".md" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
