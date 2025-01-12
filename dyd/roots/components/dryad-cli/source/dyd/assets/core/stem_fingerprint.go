package core

import (
	fs2 "dryad/filesystem"
	"io"
	"os"

	"encoding/hex"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"

	"dryad/task"

	"golang.org/x/crypto/blake2b"
)

var RE_STEM_FINGERPRINT_SHOULD_MATCH = regexp.MustCompile(
	"^(" +
		"(dyd/path/.*)" +
		"|(dyd/assets/.*)" +
		"|(dyd/commands/.*)" +
		"|(dyd/docs/.*)" +
		"|(dyd/type)" +
		"|(dyd/secrets-fingerprint)" +
		"|(dyd/traits/.*)" +
		"|(dyd/requirements/.*)" +
		")$",
)

func StemFingerprintShouldMatch(node fs2.Walk5Node) (bool, error) {
	var relPath, relErr = filepath.Rel(node.BasePath, node.VPath)
	if relErr != nil {
		return false, relErr
	}
	matchesPath := RE_STEM_FINGERPRINT_SHOULD_MATCH.Match([]byte(relPath))

	if !matchesPath {
		return false, nil
	} else if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
		linkTarget, err := os.Readlink(node.Path)
		if err != nil {
			return false, err
		}

		// clean up relative links
		if !filepath.IsAbs(linkTarget) {
			linkTarget = filepath.Clean(filepath.Join(filepath.Dir(node.Path), linkTarget))
		}

		isDescendant, err := fileIsDescendant(linkTarget, node.BasePath)
		if err != nil {
			return false, err
		}

		return isDescendant, nil
	} else if node.Info.IsDir() {
		return false, nil
	} else {
		return true, nil
	}

}

type StemFingerprintArgs struct {
	BasePath  string
	MatchDeny *regexp.Regexp
}

func StemFingerprint(args StemFingerprintArgs) (string, error) {
	var checksumMap = make(map[string]string)
	var checksumMutex sync.Mutex

	var onMatch = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
		var relPath, relErr = filepath.Rel(node.BasePath, node.VPath)
		if relErr != nil {
			return relErr, nil
		}

		if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
			var _, hash, hashErr = linkHash(node.VPath)

			if hashErr != nil {
				return hashErr, nil
			}

			checksumMutex.Lock()
			checksumMap[relPath] = hash
			checksumMutex.Unlock()
		} else {
			var _, hash, hashErr = fileHash(node.VPath)

			if hashErr != nil {
				return hashErr, nil
			}

			checksumMutex.Lock()
			checksumMap[relPath] = hash
			checksumMutex.Unlock()
		}

		return nil, nil
	}

	err, _ := fs2.BFSWalk3(
		task.DEFAULT_CONTEXT,
		fs2.Walk5Request{
			Path:        args.BasePath,
			VPath:       args.BasePath,
			BasePath:    args.BasePath,
			ShouldCrawl: StemWalkShouldCrawl,
			ShouldMatch: StemFingerprintShouldMatch,
			OnMatch:     onMatch,
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

	var checksumString = strings.Join(checksumTable, "\u0000")
	// log.Print("checksumString ", checksumString)

	hash, err := blake2b.New(16, []byte{})
	if err != nil {
		return "", err
	}

	_, err = io.WriteString(hash, "stem\u0000")
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

	// fmt.Println("StemFingerprint", args.BasePath, fingerprint)

	return fingerprint, nil
}
