package core

import (
	"fmt"
	"path/filepath"
)

func ScopeSettingPath(basePath string, scope string, setting string) (string, error) {
	scopePath, err := ScopePath(basePath, scope)
	if err != nil {
		return "", err
	}
	scopeExists, err := fileExists(scopePath)
	if err != nil {
		return "", err
	}
	if !scopeExists {
		return "", fmt.Errorf("scope %s not found", scope)
	}

	settingPath := filepath.Join(scopePath, setting)
	settingExists, err := fileExists(settingPath)
	if err != nil {
		return "", err
	}
	if !settingExists {
		return "", nil
	}

	return settingPath, nil
}
