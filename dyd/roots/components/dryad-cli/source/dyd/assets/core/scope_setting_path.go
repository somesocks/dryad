package core

import (
	"path/filepath"
)

func ScopeSettingPath(garden *SafeGardenReference, scope string, setting string) (string, error) {
	scopePath, err := ScopePath(garden, scope)
	if err != nil {
		return "", err
	}

	settingPath := filepath.Join(scopePath, setting)

	return settingPath, nil
}
