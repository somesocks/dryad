package core

import (
	"dryad/internal/os"
	"io"

	"golang.org/x/crypto/blake2b"
)

func fileHash(filePath string) (string, string, error) {
	var hashString string

	file, err := os.Open(filePath)
	if err != nil {
		return fingerprintVersionV2, hashString, err
	}
	defer file.Close()

	hash, err := blake2b.New(fingerprintDigestLen, []byte{})
	if err != nil {
		return fingerprintVersionV2, hashString, err
	}

	_, err = io.WriteString(hash, "file\u0000")
	if err != nil {
		return fingerprintVersionV2, hashString, err
	}

	_, err = io.Copy(hash, file)
	if err != nil {
		return fingerprintVersionV2, hashString, err
	}

	hashInBytes := hash.Sum(nil)[:fingerprintDigestLen]
	hashString = fingerprintEncode(hashInBytes)
	return fingerprintVersionV2, hashString, nil
}
