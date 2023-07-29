package core

import (
	"os"
	"path/filepath"
)

func ScopeSetDefault(path string, scope string) error {
	scopesPath, err := ScopesPath(path)
	if err != nil {
		return err
	}

	scopePath, err := ScopePath(path, scope)
	if err != nil {
		return err
	}
	if _, err := os.Lstat(scopePath); err != nil {
		return err
	}

	linkPath, err := filepath.Rel(scopesPath, scopePath)
	if err != nil {
		return err
	}

	defaultScopeAlias := filepath.Join(scopesPath, "default")
	if _, err := os.Lstat(defaultScopeAlias); err == nil {
		os.Remove(defaultScopeAlias)
	}
	err = os.Symlink(linkPath, defaultScopeAlias)
	if err != nil {
		return err
	}

	return nil
}
