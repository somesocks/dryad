package core

import (
	fs2 "dryad/filesystem"

	"encoding/hex"
	"io/fs"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"golang.org/x/crypto/blake2b"
)

var RE_STEM_FINGERPRINT_SHOULD_MATCH = regexp.MustCompile(
	"^(" +
		"(dyd/path/.*)" +
		"|(dyd/assets/.*)" +
		"|(dyd/assets/.*)" +
		"|(dyd/readme)" +
		"|(dyd/type)" +
		"|(dyd/main)" +
		"|(dyd/secrets-fingerprint)" +
		"|(dyd/stems/.*/dyd/fingerprint)" +
		"|(dyd/stems/.*/dyd/traits/.*)" +
		"|(dyd/stems/.*/dyd/traits/.*)" +
		"|(dyd/traits/.*)" +
		")$",
)

func StemFingerprintShouldMatch(path string, info fs.FileInfo, basePath string) (bool, error) {
	var relPath, relErr = filepath.Rel(basePath, path)
	if relErr != nil {
		return false, relErr
	}
	matchesPath := RE_STEM_FINGERPRINT_SHOULD_MATCH.Match([]byte(relPath))
	shouldMatch := matchesPath
	return shouldMatch, nil
}

type StemFingerprintArgs struct {
	BasePath  string
	MatchDeny *regexp.Regexp
}

func StemFingerprint(args StemFingerprintArgs) (string, error) {
	var checksumMap = make(map[string]string)

	var onMatch = func(walk string, info fs.FileInfo, basePath string) error {
		var rel, relErr = filepath.Rel(basePath, walk)

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

	err := fs2.BFSWalk(fs2.Walk3Request{
		BasePath:    args.BasePath,
		ShouldCrawl: StemWalkShouldCrawl,
		ShouldMatch: StemFingerprintShouldMatch,
		OnMatch:     onMatch,
	})

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
