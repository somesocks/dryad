package core

import (
	"dryad/internal/filepath"
)

func HeapPath(garden *SafeGardenReference) (string, error) {
	heapPath := filepath.Join(garden.BasePath, "dyd", "heap")
	return heapPath, nil
}
