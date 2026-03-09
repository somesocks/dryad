package core

import (
	"dryad/internal/filepath"
)

func RootsPath(garden *SafeGardenReference) (string, error) {
	return filepath.Join(garden.BasePath, "dyd", "roots"), nil
}
