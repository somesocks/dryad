package core

import (
	"path/filepath"
)

func ScopePath(garden *SafeGardenReference, scope string) (string, error) {
	var scopesPath, err = ScopesPath(garden)
	if err != nil {
		return "", err
	}

	var scopePath = filepath.Join(scopesPath, scope)
	return scopePath, nil
}
