package core

import (
	"path/filepath"
)

func HeapPath(garden *SafeGardenReference) (string, error) {
	heapPath := filepath.Join(garden.BasePath, "dyd", "heap")
	return heapPath, nil
}
