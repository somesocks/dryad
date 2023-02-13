package core

import (
	"os"
	"path/filepath"
)

func ScopeSettingUnset(basePath string, scope string, setting string) error {
	scopePath, err := ScopePath(basePath, scope)
	if err != nil {
		return err
	}

	settingPath := filepath.Join(scopePath, setting)
	if err != nil {
		return err
	}

	err = os.Remove(settingPath)

	return err
}
