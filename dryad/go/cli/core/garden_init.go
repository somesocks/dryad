package core

import (
	"log"
	"os"
	"path/filepath"
)

func GardenInit(path string) {
	var gardenPath string = filepath.Join(path, "dyd")
	if err := os.MkdirAll(gardenPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	var configPath string = filepath.Join(gardenPath, "config")
	if err := os.MkdirAll(configPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	var heapPath string = filepath.Join(gardenPath, "heap")
	if err := os.MkdirAll(heapPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	var rootsPath string = filepath.Join(gardenPath, "roots")
	if err := os.MkdirAll(rootsPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

	var stemsPath string = filepath.Join(gardenPath, "stems")
	if err := os.MkdirAll(stemsPath, os.ModePerm); err != nil {
		log.Fatal(err)
	}

}
