package core

import (
	"dryad/internal/os"
	"io"

	"golang.org/x/crypto/blake2b"
)

func linkHash(filePath string) (string, string, error) {
	var hashString string

	target, err := os.Readlink(filePath)
	if err != nil {
		return fingerprintVersionV2, hashString, err
	}

	hash, err := blake2b.New(fingerprintDigestLen, []byte{})
	if err != nil {
		return fingerprintVersionV2, hashString, err
	}

	_, err = io.WriteString(hash, "link\u0000")
	if err != nil {
		return fingerprintVersionV2, hashString, err
	}

	_, err = io.WriteString(hash, target)
	if err != nil {
		return fingerprintVersionV2, hashString, err
	}

	hashInBytes := hash.Sum(nil)[:fingerprintDigestLen]
	hashString = fingerprintEncode(hashInBytes)
	return fingerprintVersionV2, hashString, nil
}
