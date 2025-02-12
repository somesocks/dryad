package core

import (
	"io/fs"
	"os"
	"path/filepath"
)

func ScopeSettingSet(garden *SafeGardenReference, scope string, setting string, value string) error {
	scopePath, err := ScopePath(garden, scope)
	if err != nil {
		return err
	}

	settingPath := filepath.Join(scopePath, setting)
	if err != nil {
		return err
	}

	err = os.WriteFile(settingPath, []byte(value), fs.ModePerm)

	return err
}
