package core

import (
	"path/filepath"
)

func ScopePath(path string, scope string) (string, error) {
	var scopesPath, err = ScopesPath(path)
	if err != nil {
		return "", err
	}

	var scopePath = filepath.Join(scopesPath, scope)
	return scopePath, nil
}
