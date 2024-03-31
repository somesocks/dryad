package core

import (
	"fmt"
	"path/filepath"
)

func ScopeSettingExists(basePath string, scope string, setting string) (bool, error) {
	scopePath, err := ScopePath(basePath, scope)
	// fmt.Println("[debug] scopePath", scopePath, err)
	if err != nil {
		return false, err
	}

	scopeExists, err := fileExists(scopePath)
	// fmt.Println("[debug] scopeExists", scopeExists, err)
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
