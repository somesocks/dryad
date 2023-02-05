package core

import (
	"encoding/hex"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/crypto/blake2b"
)

var STEM_FINGERPRINT_MATCH_ALLOW, _ = regexp.Compile(`^((dyd/path/.*)|(dyd/assets/.*)|(dyd/readme)|(dyd/main)|(dyd/stems/.*/dyd/fingerprint)|(dyd/stems/.*/dyd/traits/.*)|(dyd/traits/.*))$`)

type StemFingerprintArgs struct {
	BasePath  string
	MatchDeny *regexp.Regexp
}

func StemFingerprint(args StemFingerprintArgs) (string, error) {
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

	err := StemWalk(
		StemWalkArgs{
			BasePath:     args.BasePath,
			MatchInclude: STEM_FINGERPRINT_MATCH_ALLOW,
			MatchExclude: args.MatchDeny,
			OnMatch:      onMatch,
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

	hash, err := blake2b.New(16, []byte{})
	if err != nil {
		return "", err
	}

	_, err = hash.Write([]byte(checksumString))
	if err != nil {
		return "", err
	}

	var fingerprintHashBytes = hash.Sum([]byte{})
	var fingerprintHash = hex.EncodeToString(fingerprintHashBytes[:])
	var fingerprint = "blake2b-" + fingerprintHash
	return fingerprint, nil
}
