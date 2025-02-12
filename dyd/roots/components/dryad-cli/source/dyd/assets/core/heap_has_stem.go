package core

import (
	"os"
	"path/filepath"
)

func HeapHasStem(garden *SafeGardenReference, fingerprint string) (string, error) {
	// fmt.Println("[trace] HeapHasStem ", path, fingerprint)

	var heapPath, heapErr = HeapPath(garden)
	// fmt.Println("HeapHasStem heapPath ", heapPath)
	if heapErr != nil {
		return "", heapErr
	}

	var stemPath = filepath.Join(heapPath, "stems", fingerprint)
	var fileInfo, fileInfoErr = os.Stat(stemPath)

	if fileInfoErr == nil && fileInfo.IsDir() {
		return stemPath, nil
	} else {
		return "", nil
	}

}
