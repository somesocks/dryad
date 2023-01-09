package core

import (
	"log"
	"os"
	"path/filepath"
)

func RootInit(path string) {
	var root_path string = filepath.Join(path, "dyd")
	if err := os.MkdirAll(root_path, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	var assets_path string = filepath.Join(root_path, "assets")
	if err := os.MkdirAll(assets_path, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	var roots_path string = filepath.Join(root_path, "roots")
	if err := os.MkdirAll(roots_path, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	var stems_path string = filepath.Join(root_path, "stems")
	if err := os.MkdirAll(stems_path, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	var main_path string = filepath.Join(root_path, "main")
	var _, main_err = os.Create(main_path)
	if main_err != nil {
		log.Fatal(main_err)
	}
	os.Chmod(main_path, 0775)

}
