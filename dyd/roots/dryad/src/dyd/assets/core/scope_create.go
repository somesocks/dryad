package core

import (
	"os"
)

func ScopeCreate(path string, scope string) (string, error) {
	var scopePath, err = ScopePath(path, scope)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(scopePath, os.ModePerm); err != nil {
		return "", err
	}

	return scopePath, nil
}
