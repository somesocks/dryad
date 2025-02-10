package core

import (
	"os"
)

func ScopeCreate(garden *SafeGardenReference, scope string) (string, error) {
	var scopePath, err = ScopePath(garden, scope)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(scopePath, os.ModePerm); err != nil {
		return "", err
	}

	return scopePath, nil
}
