package core

import (
	"dryad/internal/filepath"
)

func ScopesPath(garden *SafeGardenReference) (string, error) {
	return filepath.Join(garden.BasePath, "dyd", "shed", "scopes"), nil
}
