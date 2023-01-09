package core

import (
	"log"
	"os"
	"path/filepath"
)

func GardenInit(path string) {
	var garden_path string = filepath.Join(path, "dyd")
	if err := os.MkdirAll(garden_path, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	var config_path string = filepath.Join(garden_path, "config")
	if err := os.MkdirAll(config_path, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	var heap_path string = filepath.Join(garden_path, "heap")
	if err := os.MkdirAll(heap_path, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	var roots_path string = filepath.Join(garden_path, "roots")
	if err := os.MkdirAll(roots_path, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	var stems_path string = filepath.Join(garden_path, "stems")
	if err := os.MkdirAll(stems_path, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	var path_path string = filepath.Join(garden_path, "path")
	if err := os.MkdirAll(path_path, os.ModePerm); err != nil {
		log.Fatal(err)
	}

}
