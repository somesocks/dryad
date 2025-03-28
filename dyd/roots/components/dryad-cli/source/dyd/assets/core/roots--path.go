package core

import (
	"path/filepath"
)

func RootsPath(garden *SafeGardenReference) (string, error) {
	return filepath.Join(garden.BasePath, "dyd", "roots"), nil
}
