package walk

import (
	"os"
	"path/filepath"
)

// Walk returns all *.md file paths found recursively under root.
func Walk(root string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && filepath.Ext(path) == ".md" {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}
