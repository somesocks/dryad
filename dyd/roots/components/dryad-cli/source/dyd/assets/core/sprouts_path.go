package core

import (
	"path/filepath"
)

func SproutsPath(garden *SafeGardenReference) (string, error) {
	return filepath.Join(garden.BasePath, "dyd", "sprouts"), nil
}
