package core

import (
	"encoding/hex"
	"io"
	"os"

	"golang.org/x/crypto/blake2b"
)

func linkHash(filePath string) (string, string, error) {
	var hashString string

	target, err := os.Readlink(filePath)
	if err != nil {
		return "dyd-v1", hashString, err
	}

	hash, err := blake2b.New(16, []byte{})
	if err != nil {
		return "dyd-v1", hashString, err
	}

	_, err = io.WriteString(hash, "link\u0000")
	if err != nil {
		return "dyd-v1", hashString, err
	}

	_, err = io.WriteString(hash, target)
	if err != nil {
		return "dyd-v1", hashString, err
	}

	hashInBytes := hash.Sum(nil)[:16]
	hashString = hex.EncodeToString(hashInBytes)
	return "dyd-v1", hashString, nil
}
