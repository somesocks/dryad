package core

import (
	"fmt"
	"path/filepath"
)

func ScopeSettingExists(garden *SafeGardenReference, scope string, setting string) (bool, error) {
	scopePath, err := ScopePath(garden, scope)
	if err != nil {
		return false, err
	}

	scopeExists, err := fileExists(scopePath)
	if err != nil {
		return false, err
	}
	if !scopeExists {
		return false, fmt.Errorf("scope %s not found", scope)
	}

	settingPath := filepath.Join(scopePath, setting)

	settingExists, err := fileExists(settingPath)

	return settingExists, err
}
