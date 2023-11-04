package core

import (
	"os"
	"path/filepath"
)

func StemInit(path string) error {
	var root_path string = filepath.Join(path, "dyd")
	if err := os.MkdirAll(root_path, os.ModePerm); err != nil {
		return err
	}

	var assets_path string = filepath.Join(root_path, "assets")
	if err := os.MkdirAll(assets_path, os.ModePerm); err != nil {
		return err
	}

	var commandsPath string = filepath.Join(root_path, "commands")
	if err := os.MkdirAll(commandsPath, os.ModePerm); err != nil {
		return err
	}

	var stems_path string = filepath.Join(root_path, "stems")
	if err := os.MkdirAll(stems_path, os.ModePerm); err != nil {
		return err
	}

	var traits_path string = filepath.Join(root_path, "traits")
	if err := os.MkdirAll(traits_path, os.ModePerm); err != nil {
		return err
	}

	return nil
}
