package core

import (
	"fmt"
	"os"
	"path/filepath"
)

func RootCreate(path string) error {
	path, err := filepath.Abs(path)
	if err != nil {
		return err
	}

	pathExists, err := fileExists(path)
	if err != nil {
		return err
	} else if pathExists {
		return fmt.Errorf("error: init destination %s already exists", path)
	}

	var basePath string = filepath.Join(path, "dyd")
	if err := os.MkdirAll(basePath, os.ModePerm); err != nil {
		return err
	}

	// write out type file
	typePath := filepath.Join(basePath, "type")

	typeFile, err := os.Create(typePath)
	if err != nil {
		return err
	}
	defer typeFile.Close()

	_, err = fmt.Fprint(typeFile, "root")
	if err != nil {
		return err
	}

	err = typeFile.Sync()
	if err != nil {
		return err
	}

	var assetsPath string = filepath.Join(basePath, "assets")
	if err := os.MkdirAll(assetsPath, os.ModePerm); err != nil {
		return err
	}

	var commandsPath string = filepath.Join(basePath, "commands")
	if err := os.MkdirAll(commandsPath, os.ModePerm); err != nil {
		return err
	}

	var rootsPath string = filepath.Join(basePath, "requirements")
	if err := os.MkdirAll(rootsPath, os.ModePerm); err != nil {
		return err
	}

	var stemsPath string = filepath.Join(basePath, "stems")
	if err := os.MkdirAll(stemsPath, os.ModePerm); err != nil {
		return err
	}

	var traitsPath string = filepath.Join(basePath, "traits")
	if err := os.MkdirAll(traitsPath, os.ModePerm); err != nil {
		return err
	}

	var secretsPath string = filepath.Join(basePath, "secrets")
	if err := os.MkdirAll(secretsPath, os.ModePerm); err != nil {
		return err
	}

	var defaultCommandPath string = filepath.Join(basePath, "commands", "default")
	if _, err := os.Create(defaultCommandPath); err != nil {
		return err
	}

	if err := os.Chmod(defaultCommandPath, 0775); err != nil {
		return err
	}

	return nil
}
