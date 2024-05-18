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
		return "dyd-v1", hashString, err
	}
	defer file.Close()

	hash, err := blake2b.New(16, []byte{})
	if err != nil {
		return "dyd-v1", hashString, err
	}

	_, err = io.WriteString(hash, "file\u0000")
	if err != nil {
		return "dyd-v1", hashString, err
	}

	_, err = io.Copy(hash, file)
	if err != nil {
		return "dyd-v1", hashString, err
	}

	hashInBytes := hash.Sum(nil)[:16]
	hashString = hex.EncodeToString(hashInBytes)
	return "dyd-v1", hashString, nil
}
