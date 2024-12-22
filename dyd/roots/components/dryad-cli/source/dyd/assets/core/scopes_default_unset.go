package core

import (
	"os"
	"path/filepath"
)

func ScopeUnsetDefault(path string) error {
	scopesPath, err := ScopesPath(path)
	if err != nil {
		return err
	}

	defaultScopeAlias := filepath.Join(scopesPath, "default")
	if _, err := os.Lstat(defaultScopeAlias); err == nil {
		os.Remove(defaultScopeAlias)
	}
	if err != nil {
		return err
	}

	return nil
}
