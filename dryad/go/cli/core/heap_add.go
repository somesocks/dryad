package core

import (
	"os"
	"path/filepath"
)

func HeapAdd(heapPath string, filePath string) (string, error) {
	heapPath, err := HeapPath(heapPath)
	if err != nil {
		return "", err
	}

	fileHashAlgorithm, fileHash, err := fileHash(filePath)
	if err != nil {
		return "", err
	}

	fingerprint := fileHashAlgorithm + "-" + fileHash

	destPath := filepath.Join(heapPath, fingerprint)

	fileExists, err := fileExists(destPath)
	if err != nil {
		return "", err
	}

	if !fileExists {
		srcFile, err := os.Open(filePath)
		if err != nil {
			return "", err
		}
		defer srcFile.Close()

		var destFile *os.File
		destFile, err = os.Create(destPath)
		if err != nil {
			return "", err
		}
		defer destFile.Close()

		_, err = destFile.ReadFrom(srcFile)
		if err != nil {
			return "", err
		}

		err = destFile.Chmod(os.ModePerm)
		if err != nil {
			return "", err
		}

		err = destFile.Sync()
		if err != nil {
			return "", err
		}
	}

	return fingerprint, nil
}

// func HeapAdd(heapPath string, stemPath string) (string, error) {
// 	var err error
// 	heapPath, err = HeapPath(heapPath)
// 	if err != nil {
// 		return "", err
// 	}

// 	var stemFingerprint string
// 	stemFingerprint, err = StemFingerprint(stemPath)
// 	if err != nil {
// 		return "", err
// 	}

// 	var destPath = filepath.Join(heapPath, stemFingerprint)

// 	// check if the stem already exists in the heap
// 	var alreadyExists bool

// 	alreadyExists, err = fileExists(destPath)
// 	if err != nil {
// 		return "", err
// 	}

// 	if alreadyExists {
// 		return stemFingerprint, nil
// 	}

// 	var workingPath = filepath.Join(heapPath, "_"+stemFingerprint)

// 	_, err = StemPack(stemPath, workingPath)
// 	if err != nil {
// 		return "", err
// 	}

// 	var dependenciesPath = filepath.Join(workingPath, "dyd", "stems")

// 	var dependencies []string
// 	dependencies, err = filepath.Glob(dependenciesPath + "/*")
// 	if err != nil {
// 		return "", err
// 	}

// 	// replace dependency stubs with symlinks
// 	for _, dependencyPath := range dependencies {
// 		var dependencyFingerprint string
// 		var dependencyFingerprintBytes []byte

// 		dependencyFingerprintBytes, err = ioutil.ReadFile(filepath.Join(dependencyPath, "dyd", "fingerprint"))
// 		if err != nil {
// 			return "", err
// 		}
// 		dependencyFingerprint = string(dependencyFingerprintBytes)

// 		err = os.RemoveAll(dependencyPath)
// 		if err != nil {
// 			return "", err
// 		}

// 		var depedencyTargetPath = filepath.Join(heapPath, dependencyFingerprint)
// 		var relativeDependencyPath string
// 		relativeDependencyPath, err = filepath.Rel(dependenciesPath, depedencyTargetPath)
// 		if err != nil {
// 			return "", err
// 		}

// 		err = os.Symlink(relativeDependencyPath, dependencyPath)
// 		if err != nil {
// 			return "", err
// 		}

// 	}

// 	err = os.Rename(workingPath, destPath)
// 	if err != nil {
// 		return "", err
// 	}

// 	return stemFingerprint, nil

// }
