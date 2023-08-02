package core

import (
	"fmt"
	"path/filepath"
)

func ScopeSettingPath(basePath string, scope string, setting string) (string, error) {
	scopePath, err := ScopePath(basePath, scope)
	// fmt.Println("[debug] scopePath", scopePath, err)
	if err != nil {
		return "", err
	}

	scopeExists, err := fileExists(scopePath)
	// fmt.Println("[debug] scopeExists", scopeExists, err)
	if err != nil {
		return "", err
	}
	if !scopeExists {
		return "", fmt.Errorf("scope %s not found", scope)
	}

	settingPath := filepath.Join(scopePath, setting)
	// fmt.Println("[debug] settingPath", settingPath)

	settingExists, err := fileExists(settingPath)
	// fmt.Println("[debug] settingExists", settingExists, err)

	if err != nil {
		return "", err
	}
	if !settingExists {
		return "", nil
	}

	return settingPath, nil
}
