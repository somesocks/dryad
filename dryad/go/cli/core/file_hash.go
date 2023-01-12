package core

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
)

// func fileHash(filePath string) (string, string, error) {
// 	hashString := ""
// 	hasAlgorithm := "blake2b"

// 	file, err := os.Open(filePath)
// 	if err != nil {
// 		return hasAlgorithm, hashString, err
// 	}
// 	defer file.Close()

// 	hash, err := blake2b.New256([]byte{})
// 	if err != nil {
// 		return hasAlgorithm, hashString, err
// 	}

// 	if _, err := io.Copy(hash, file); err != nil {
// 		return hasAlgorithm, hashString, err
// 	}
// 	//Get the 16 bytes hash
// 	hashInBytes := hash.Sum(nil)[:16]
// 	//Convert the bytes to a string
// 	hashString = hex.EncodeToString(hashInBytes)
// 	return hasAlgorithm, hashString, nil
// }

func fileHash(filePath string) (string, string, error) {
	//Initialize variable returnMD5String now in case an error has to be returned
	var returnMD5String string
	//Open the passed argument and check for any error
	file, err := os.Open(filePath)
	if err != nil {
		return "md5sum", returnMD5String, err
	}
	//Tell the program to call the following function when the current function returns
	defer file.Close()
	//Open a new hash interface to write to
	hash := md5.New()
	//Copy the file in the hash interface and check for any error
	if _, err := io.Copy(hash, file); err != nil {
		return "md5sum", returnMD5String, err
	}
	//Get the 16 bytes hash
	hashInBytes := hash.Sum(nil)[:16]
	//Convert the bytes to a string
	returnMD5String = hex.EncodeToString(hashInBytes)
	return "md5sum", returnMD5String, nil
}
