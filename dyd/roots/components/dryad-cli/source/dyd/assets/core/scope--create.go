package core

import (
	"dryad/internal/os"
	stdos "os"
)

func ScopeCreate(garden *SafeGardenReference, scope string) (string, error) {
	var scopePath, err = ScopePath(garden, scope)
	if err != nil {
		return "", err
	}

	if err := os.MkdirAll(scopePath, stdos.ModePerm); err != nil {
		return "", err
	}

	return scopePath, nil
}
