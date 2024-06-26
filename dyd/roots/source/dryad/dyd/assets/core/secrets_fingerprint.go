package core

import (
	"encoding/hex"
	"io"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/crypto/blake2b"
)

type SecretsFingerprintArgs struct {
	BasePath string
}

func SecretsFingerprint(args SecretsFingerprintArgs) (string, error) {
	var checksumMap = make(map[string]string)

	var onMatch = func(walk string, info fs.FileInfo) error {
		var rel, relErr = filepath.Rel(args.BasePath, walk)

		if relErr != nil {
			return relErr
		}

		if info.IsDir() {
			return nil
		}

		var _, hash, hashErr = fileHash(walk)

		if hashErr != nil {
			return hashErr
		}

		checksumMap[rel] = hash

		return nil
	}

	err := SecretsWalk(
		SecretsWalkArgs{
			BasePath: args.BasePath,
			OnMatch:  onMatch,
		},
	)
	if err != nil {
		return "", err
	}

	// if there are no secrets,
	// return an empty fingerprint
	if len(checksumMap) == 0 {
		return "", nil
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

	var checksumString = strings.Join(checksumTable, "\u0000")
	// log.Print("checksumString ", checksumString)

	hash, err := blake2b.New(16, []byte{})
	if err != nil {
		return "", err
	}

	_, err = io.WriteString(hash, "secrets\u0000")
	if err != nil {
		return "", err
	}

	_, err = io.WriteString(hash, checksumString)
	if err != nil {
		return "", err
	}

	var fingerprintHashBytes = hash.Sum([]byte{})
	var fingerprintHash = hex.EncodeToString(fingerprintHashBytes[:])
	var fingerprint = "dyd-v1-" + fingerprintHash

	return fingerprint, nil
}
