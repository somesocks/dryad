package core

import (
	"dryad/internal/os"
	stdos "os"
	"path/filepath"
)

func StemInit(path string) error {
	var root_path string = filepath.Join(path, "dyd")
	if err := os.MkdirAll(root_path, stdos.ModePerm); err != nil {
		return err
	}

	var assets_path string = filepath.Join(root_path, "assets")
	if err := os.MkdirAll(assets_path, stdos.ModePerm); err != nil {
		return err
	}

	var commandsPath string = filepath.Join(root_path, "commands")
	if err := os.MkdirAll(commandsPath, stdos.ModePerm); err != nil {
		return err
	}

	var docsPath string = filepath.Join(root_path, "docs")
	if err := os.MkdirAll(docsPath, stdos.ModePerm); err != nil {
		return err
	}

	var stems_path string = filepath.Join(root_path, "dependencies")
	if err := os.MkdirAll(stems_path, stdos.ModePerm); err != nil {
		return err
	}

	var traits_path string = filepath.Join(root_path, "traits")
	if err := os.MkdirAll(traits_path, stdos.ModePerm); err != nil {
		return err
	}

	return nil
}
