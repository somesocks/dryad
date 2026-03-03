package core

import (
	"dryad/internal/os"
	stdos "os"
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
	if _, err := stdos.Lstat(scopePath); err != nil {
		return err
	}

	linkPath, err := filepath.Rel(scopesPath, scopePath)
	if err != nil {
		return err
	}

	defaultScopeAlias := filepath.Join(scopesPath, "default")
	if _, err := stdos.Lstat(defaultScopeAlias); err == nil {
		stdos.Remove(defaultScopeAlias)
	}
	err = os.Symlink(linkPath, defaultScopeAlias)
	if err != nil {
		return err
	}

	return nil
}
