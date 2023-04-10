package core

import (
	"log"
	"os"
	"path/filepath"
)

func HeapPath(path string) (string, error) {
	// fmt.Println("[trace] HeapPath " + path)

	gardenPath, err := GardenPath(path)

	if err != nil {
		log.Fatal(err)
	}

	heapPath := filepath.Join(gardenPath, "dyd", "heap")
	fileInfo, err := os.Stat(heapPath)
	if err != nil {
		return "", err
	}

	if !fileInfo.IsDir() {
		err = GardenInit(gardenPath)
	}
	if err != nil {
		return "", err
	}

	return heapPath, nil
}
