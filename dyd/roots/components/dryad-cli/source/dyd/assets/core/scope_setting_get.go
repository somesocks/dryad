package core

import (
	"dryad/internal/filepath"
	"dryad/internal/os"
	"fmt"
)

func ScopeSettingGet(garden *SafeGardenReference, scope string, setting string) (string, error) {
	scopePath, err := ScopePath(garden, scope)
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

	settingBytes, err := os.ReadFile(settingPath)
	if err != nil {
		return "", err
	}

	settingString := string(settingBytes)

	return settingString, nil
}
