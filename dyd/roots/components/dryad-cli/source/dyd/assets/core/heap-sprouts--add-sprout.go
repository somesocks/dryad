package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"errors"
	"io/fs"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

type heapAddSproutRequest struct {
	HeapSprouts *SafeHeapSproutsReference
	HeapFiles   *SafeHeapFilesReference
	SproutPath  string
}

// heapAddSprout takes a sprout in a directory and adds it to the heap.
// the heap path is normalized before adding
func heapAddSprout(ctx *task.ExecutionContext, req heapAddSproutRequest) (error, *SafeHeapSproutReference) {
	sproutPath := req.SproutPath

	heapFilesPath := req.HeapFiles.BasePath
	heapSproutsPath := req.HeapSprouts.BasePath
	heapStemsPath := filepath.Join(req.HeapSprouts.Heap.BasePath, "stems")

	sproutFingerprint, err := _readFile(filepath.Join(sproutPath, "dyd", "fingerprint"))
	if err != nil {
		return err, nil
	}

	finalSproutPath := filepath.Join(heapSproutsPath, sproutFingerprint)

	// check to see if the sprout already exists in the garden
	sproutExists, err := fileExists(finalSproutPath)
	if err != nil {
		return err, nil
	}

	// if sprout exists, do nothing
	if sproutExists {
		sproutRef := SafeHeapSproutReference{
			BasePath: finalSproutPath,
			Sprouts:  req.HeapSprouts,
		}

		return nil, &sproutRef
	}

	err = os.MkdirAll(finalSproutPath, fs.ModePerm)
	if err != nil {
		return err, nil
	}

	// walk the packed sprout files and copy them into the garden heap
	err, _ = StemWalk(
		ctx,
		StemWalkRequest{
			BasePath: sproutPath,
			OnMatch: func(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, any) {
				zlog.
					Trace().
					Str("path", node.Path).
					Str("vpath", node.VPath).
					Msg("heapAddSprout / onMatch")

				relPath, err := filepath.Rel(node.BasePath, node.VPath)
				if err != nil {
					return err, nil
				}

				destPath := filepath.Join(finalSproutPath, relPath)

				// if the file already exists, we hit it on a previous pass through a symlink
				destExists, err := fileExists(destPath)
				if err != nil {
					return err, nil
				}
				if destExists {
					return errors.New("heap add sprout error - file already exists but should not"), nil
				}

				if node.Info.IsDir() {
					err = os.Mkdir(destPath, os.ModePerm)
					if err != nil {
						return err, nil
					}
				} else if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
					linkTarget, err := os.Readlink(node.Path)
					if err != nil {
						return err, nil
					}

					absLinkTarget := linkTarget
					if !filepath.IsAbs(absLinkTarget) {
						absLinkTarget = filepath.Join(filepath.Dir(node.VPath), linkTarget)
					}

					isInternalLink, err := fileIsDescendant(absLinkTarget, node.BasePath)

					if isInternalLink {
						err = os.Symlink(linkTarget, destPath)
						if err != nil {
							return err, nil
						}
					}
				} else {
					err, fileFingerprint := req.HeapFiles.AddFile(
						ctx,
						node.Path,
					)
					if err != nil {
						zlog.
							Trace().
							Str("path", node.Path).
							Str("vpath", node.VPath).
							Err(err).
							Msg("heapAddSprout / onMatch / HeapAddFile error")
						return err, nil
					}

					fileHeapPath := filepath.Join(heapFilesPath, fileFingerprint)

					err = os.Link(fileHeapPath, destPath)
					if err != nil {
						return err, nil
					}
				}

				return nil, nil
			},
		},
	)
	if err != nil {
		return err, nil
	}

	// rebuild dependency links from the source sprout dependencies.
	sourceDependenciesPath := filepath.Join(sproutPath, "dyd", "dependencies")
	dependenciesPath := filepath.Join(finalSproutPath, "dyd", "dependencies")
	dependencies, err := filepath.Glob(filepath.Join(sourceDependenciesPath, "*"))
	if err != nil {
		return err, nil
	}

	for _, dependencySourcePath := range dependencies {
		targetStemPath, err := filepath.EvalSymlinks(dependencySourcePath)
		if err != nil {
			return err, nil
		}

		targetFingerprintFile := filepath.Join(targetStemPath, "dyd", "fingerprint")
		targetFingerprintBytes, err := os.ReadFile(targetFingerprintFile)
		if err != nil {
			return err, nil
		}
		targetFingerprint := string(targetFingerprintBytes)

		dependencyPath := filepath.Join(dependenciesPath, filepath.Base(dependencySourcePath))
		dependencyGardenPath := filepath.Join(heapStemsPath, targetFingerprint)
		relPath, err := filepath.Rel(dependenciesPath, dependencyGardenPath)
		if err != nil {
			return err, nil
		}

		err = os.Symlink(relPath, dependencyPath)
		if err != nil {
			return err, nil
		}
	}

	setPermissionsShouldCrawl := func(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, bool) {
		isDir := node.Info.IsDir()

		zlog.Trace().
			Str("path", node.VPath).
			Bool("shouldCrawl", isDir).
			Msg("heap add sprout - dir ShouldCrawl")

		return nil, isDir
	}

	setPermissionsShouldMatch := func(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, bool) {
		isDir := node.Info.IsDir()

		zlog.Trace().
			Str("path", node.VPath).
			Bool("shouldMatch", isDir).
			Msg("heap add sprout - dir ShouldMatch")

		return nil, isDir
	}

	setPermissionsOnMatch := func(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, any) {
		zlog.Trace().
			Str("path", node.VPath).
			Msg("heap add sprout - dir OnMatch")

		dirPerms := node.Info.Mode().Perm()

		// if permissions are already set correctly, do nothing
		if dirPerms == 0o511 {
			return nil, nil
		}

		dir, err := os.Open(node.Path)
		if err != nil {
			return err, nil
		}
		defer dir.Close()

		// heap files should be set to R-X--X--X
		err = dir.Chmod(0o511)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	// now that all files are added, sweep through in a second pass and make directories read-only
	err, _ = fs2.Walk6(
		ctx,
		fs2.Walk6Request{
			BasePath:    finalSproutPath,
			Path:        finalSproutPath,
			VPath:       finalSproutPath,
			ShouldWalk:  setPermissionsShouldCrawl,
			OnPostMatch: fs2.ConditionalWalkAction(setPermissionsOnMatch, setPermissionsShouldMatch),
		},
	)
	if err != nil {
		return err, nil
	}

	sproutRef := SafeHeapSproutReference{
		BasePath: finalSproutPath,
		Sprouts:  req.HeapSprouts,
	}

	return nil, &sproutRef
}

var memoHeapAddSprout = task.Memoize(
	heapAddSprout,
	func(ctx *task.ExecutionContext, req heapAddSproutRequest) (error, any) {
		type Key struct {
			Group       string
			Fingerprint string
		}
		var res Key
		var fingerprint string
		var err error

		fingerprint, err = _readFile(
			filepath.Join(req.SproutPath, "dyd", "fingerprint"),
		)
		if err != nil {
			return err, res
		}

		res = Key{
			Group:       "HeapSprouts.AddSprout",
			Fingerprint: fingerprint,
		}

		return nil, res
	},
)

type HeapAddSproutRequest struct {
	SproutPath string
}

func (heapSprouts *SafeHeapSproutsReference) AddSprout(
	ctx *task.ExecutionContext,
	req HeapAddSproutRequest,
) (error, *SafeHeapSproutReference) {
	err, heapFiles := heapSprouts.Heap.Files().Resolve(ctx)
	if err != nil {
		return err, nil
	}

	err, res := memoHeapAddSprout(
		ctx,
		heapAddSproutRequest{
			HeapSprouts: heapSprouts,
			HeapFiles:   heapFiles,
			SproutPath:  req.SproutPath,
		},
	)

	return err, res
}
