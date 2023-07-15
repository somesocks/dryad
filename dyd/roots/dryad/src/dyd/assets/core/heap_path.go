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
	_, err = os.Stat(heapPath)
	if err != nil {
		err = GardenCreate(gardenPath)
	}

	if err != nil {
		return "", err
	}

	return heapPath, nil
}
