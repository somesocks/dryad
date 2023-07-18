package core

import (
	fs2 "dryad/filesystem"

	"encoding/hex"
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

func StemFingerprintShouldMatch(context fs2.Walk4Context) (bool, error) {
	var relPath, relErr = filepath.Rel(context.BasePath, context.VPath)
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

	var onMatch = func(context fs2.Walk4Context) error {
		// fmt.Println("StemFingerprint onMatch", context)

		var rel, relErr = filepath.Rel(context.BasePath, context.VPath)

		if relErr != nil {
			return relErr
		}

		if context.Info.IsDir() {
			return nil
		}

		var _, hash, hashErr = fileHash(context.VPath)

		if hashErr != nil {
			return hashErr
		}

		checksumMap[rel] = hash

		return nil
	}

	err := fs2.BFSWalk2(fs2.Walk4Request{
		Path:        args.BasePath,
		VPath:       args.BasePath,
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

	// fmt.Println("StemFingerprint", args.BasePath, fingerprint)

	return fingerprint, nil
}
