package core

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func hash_file_md5(filePath string) (string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string
	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	//Tell the program to call the following function when the current function returns
	defer file.Close()
	//Open a new hash interface to write to
	hash := md5.New()
	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]
	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil
}

func StemFingerprint(path string) (string, error) {
	var checksumMap = make(map[string]string)

	var onMatch = func(walk string, info fs.FileInfo, err error) error {
		var rel, relErr = filepath.Rel(path, walk)

		if relErr != nil {
			return relErr
		}

		if info.IsDir() {
			return nil
		}

		var hash, hashErr = hash_file_md5(walk)

		if hashErr != nil {
			return hashErr
		}

		checksumMap[rel] = hash

		// fmt.Println("StemFingerprint ", path, " ", rel, " ", hash)

		return nil
	}

	err := StemWalk(
		StemWalkArgs{
			BasePath: path,
			OnMatch:  onMatch,
		},
	)
	if err != nil {
		return "", err
	}

	var keys []string
	for key, _ := range checksumMap {
		keys = append(keys, key)
	}

	sort.Strings(keys)

	var checksumTable []string

	for _, key := range keys {
		checksumTable = append(checksumTable, checksumMap[key]+" ./"+key)
	}

	var checksumString = strings.Join(checksumTable, " ")
	// log.Print("checksumString ", checksumString)

	var fingerprintHashBytes = md5.Sum([]byte(checksumString))
	var fingerprintHash = hex.EncodeToString(fingerprintHashBytes[:])
	var fingerprint = "md5sum-" + fingerprintHash
	// fmt.Printf("Key: %d, Value: %s\n", key, checksumMap[key])
	// log.Print(checksumTable)
	return fingerprint, nil
}
