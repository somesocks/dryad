package core

import (
	"path/filepath"
)

func ScopeSettingPath(basePath string, scope string, setting string) (string, error) {
	scopePath, err := ScopePath(basePath, scope)
	// fmt.Println("[debug] scopePath", scopePath, err)
	if err != nil {
		return "", err
	}

	settingPath := filepath.Join(scopePath, setting)

	return settingPath, nil
}
