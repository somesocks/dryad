package core

import (
	"os"
	"path/filepath"
)

func ScopeSetDefault(garden *SafeGardenReference, scope string) error {
	scopesPath, err := ScopesPath(garden)
	if err != nil {
		return err
	}

	scopePath, err := ScopePath(garden, scope)
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
