package core

import (
	"encoding/hex"
	"io"
	"os"

	"golang.org/x/crypto/blake2b"
)

func fileHash(filePath string) (string, string, error) {
	var hashString string

	file, err := os.Open(filePath)
	if err != nil {
		return "blake2b", hashString, err
	}
	defer file.Close()

	hash, err := blake2b.New(16, []byte{})
	if err != nil {
		return "blake2b", hashString, err
	}

	if _, err := io.Copy(hash, file); err != nil {
		return "blake2b", hashString, err
	}

	hashInBytes := hash.Sum(nil)[:16]
	hashString = hex.EncodeToString(hashInBytes)
	return "blake2b", hashString, nil
}
