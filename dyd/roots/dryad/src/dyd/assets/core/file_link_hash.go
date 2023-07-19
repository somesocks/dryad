package core

import (
	"encoding/hex"
	"os"

	"golang.org/x/crypto/blake2b"
)

func linkHash(filePath string) (string, string, error) {
	var hashString string

	target, err := os.Readlink(filePath)
	if err != nil {
		return "blake2b", hashString, err
	}

	hash, err := blake2b.New(16, []byte{})
	if err != nil {
		return "blake2b", hashString, err
	}

	_, err = hash.Write([]byte(target))
	if err != nil {
		return "blake2b", hashString, err
	}

	hashInBytes := hash.Sum(nil)[:16]
	hashString = hex.EncodeToString(hashInBytes)
	return "blake2b", hashString, nil
}
