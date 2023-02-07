package core

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
)

func ScopeResolve(basePath string, command string) string, error {
	gardenPath, err := GardenPath(basePath)

	return "", err
}
