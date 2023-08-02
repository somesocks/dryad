package core

import (
	"os"
	"path/filepath"
)

func ScopeGetDefault(path string) (string, error) {
	scopesPath, err := ScopesPath(path)
	if err != nil {
		return "", err
	}

	defaultScopeAlias := filepath.Join(scopesPath, "default")
	defaultExists, err := fileExists(defaultScopeAlias)
	if err != nil {
		return "", err
	}

	if !defaultExists {
		return "", nil
	}

	scopePath, err := os.Readlink(defaultScopeAlias)
	if err != nil {
		return "", err
	}

	scopeName := filepath.Base(scopePath)

	return scopeName, nil
}
