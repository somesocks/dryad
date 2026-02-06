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
		"|(dyd/secrets/.*)" +
		"|(dyd/commands/.*)" +
		"|(dyd/docs/.*)" +
		"|(dyd/type)" +
		"|(dyd/traits/.*)" +
		"|(dyd/requirements/.*)" +
		")$",
)

func StemFingerprintShouldMatch(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, bool) {
	var relPath, relErr = filepath.Rel(node.BasePath, node.VPath)
	if relErr != nil {
		return relErr, false
	}
	matchesPath := RE_STEM_FINGERPRINT_SHOULD_MATCH.Match([]byte(relPath))

	if !matchesPath {
		return nil, false
	} else if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
		linkTarget, err := os.Readlink(node.Path)
		if err != nil {
			return err, false
		}

		// Resolve relative link targets against the virtual path.
		// During root build stage 0 we walk through symlinked directories,
		// so node.Path may point outside BasePath even for package-internal links.
		if !filepath.IsAbs(linkTarget) {
			linkTarget = filepath.Clean(filepath.Join(filepath.Dir(node.VPath), linkTarget))
		}

		isDescendant, err := fileIsDescendant(linkTarget, node.BasePath)
		if err != nil {
			return err, false
		}

		return nil, isDescendant
	} else if node.Info.IsDir() {
		return nil, false
	} else {
		return nil, true
	}

}

type StemFingerprintRequest struct {
	BasePath  string
	MatchDeny *regexp.Regexp
}

var StemFingerprint task.Task[StemFingerprintRequest, string] = func() task.Task[StemFingerprintRequest, string] {
	var stemFingerprint = func(ctx *task.ExecutionContext, args StemFingerprintRequest) (error, string) {
		var checksumMap = make(map[string]string)
		var checksumMutex sync.Mutex

		var onMatch = func(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, any) {
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

		err, _ := fs2.Walk6(
			ctx,
			fs2.Walk6Request{
				Path:       args.BasePath,
				VPath:      args.BasePath,
				BasePath:   args.BasePath,
				ShouldWalk: StemWalkShouldCrawl,
				OnPreMatch: fs2.ConditionalWalkAction(onMatch, StemFingerprintShouldMatch),
			},
		)

		if err != nil {
			return err, ""
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
			return err, ""
		}

		_, err = io.WriteString(hash, "stem\u0000")
		if err != nil {
			return err, ""
		}

		_, err = io.WriteString(hash, checksumString)
		if err != nil {
			return err, ""
		}

		var fingerprintHashBytes = hash.Sum([]byte{})
		var fingerprintHash = hex.EncodeToString(fingerprintHashBytes[:])
		var fingerprint = "dyd-v1-" + fingerprintHash

		// fmt.Println("StemFingerprint", args.BasePath, fingerprint)

		return nil, fingerprint
	}

	// we want to replace the execution context, but with the same concurrency channel as before.
	// only the execution cache is replaced, to limit the scope of memoized calls to fetch dyd-ignore files
	stemFingerprint = task.WithContext(
		stemFingerprint,
		func(ctx *task.ExecutionContext, args StemFingerprintRequest) (error, *task.ExecutionContext) {
			return nil, &task.ExecutionContext{
				ConcurrencyChannel: ctx.ConcurrencyChannel,
			}
		},
	)

	return stemFingerprint
}()
