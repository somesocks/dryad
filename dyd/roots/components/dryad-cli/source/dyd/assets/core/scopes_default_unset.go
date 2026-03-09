package core

import (
	"dryad/internal/filepath"
	"dryad/internal/os"
)

func ScopeUnsetDefault(garden *SafeGardenReference) error {
	scopesPath, err := ScopesPath(garden)
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
