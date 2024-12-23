package core

import (
	"os"
	"path/filepath"
	"strings"
)

func fileIsDescendant(path string, parent string) (bool, error) {
	up := ".." + string(os.PathSeparator)

	// path-comparisons using filepath.Abs don't work reliably according to docs (no unique representation).
	rel, err := filepath.Rel(parent, path)
	if err != nil {
		return false, err
	}
	if !strings.HasPrefix(rel, up) && rel != ".." && rel != "." {
		return true, nil
	}
	return false, nil
}
