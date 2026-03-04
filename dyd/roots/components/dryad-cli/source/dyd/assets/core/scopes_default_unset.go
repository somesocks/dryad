package core

import (
	"dryad/internal/os"
	stdos "os"
	"path/filepath"
)

func ScopeUnsetDefault(garden *SafeGardenReference) error {
	scopesPath, err := ScopesPath(garden)
	if err != nil {
		return err
	}

	defaultScopeAlias := filepath.Join(scopesPath, "default")
	if _, err := stdos.Lstat(defaultScopeAlias); err == nil {
		os.Remove(defaultScopeAlias)
	}
	if err != nil {
		return err
	}

	return nil
}
