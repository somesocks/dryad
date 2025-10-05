package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"errors"

	zlog "github.com/rs/zerolog/log"
)

func _readFile(filePath string) (string, error) {
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

type heapAddStemRequest struct {
	HeapStems *SafeHeapStemsReference
	HeapFiles *SafeHeapFilesReference
	StemPath string	
}

// heapAddStem takes a stem in a directory, and adds it to the heap.
// the heap path is normalized before adding
func heapAddStem(ctx *task.ExecutionContext, req heapAddStemRequest) (error, *SafeHeapStemReference) {
	var stemPath = req.StemPath

	heapFilesPath := req.HeapFiles.BasePath
	heapStemsPath := req.HeapStems.BasePath

	stemFingerprint, err := _readFile(filepath.Join(stemPath, "dyd", "fingerprint"))
	if err != nil {
		return err, nil
	}

	finalStemPath := filepath.Join(heapStemsPath, stemFingerprint)

	// check to see if the stem already exists in the garden
	stemExists, err := fileExists(finalStemPath)
	if err != nil {
		return err, nil
	}

	// if stem exists, do nothing
	if stemExists {
		stemRef := SafeHeapStemReference{
			BasePath: finalStemPath,
			Stems: req.HeapStems,
		}
		
		return nil, &stemRef
	}

	err = os.MkdirAll(finalStemPath, fs.ModePerm)
	if err != nil {
		return err, nil
	}

	// walk the packed root files and copy them into the garden heap
	err, _ = StemWalk(
		ctx,
		StemWalkRequest{
			BasePath: stemPath,
			OnMatch: func(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, any) {
				zlog.
					Trace().
					Str("path", node.Path).
					Str("vpath", node.VPath).
					Msg("heapAddStem / onMatch")

				relPath, err := filepath.Rel(node.BasePath, node.VPath)
				if err != nil {
					return err, nil
				}

				destPath := filepath.Join(finalStemPath, relPath)

				// if the file already exists, we hit it on a previous pass through a symlink
				destExists, err := fileExists(destPath)
				if err != nil {
					return err, nil
				}
				if destExists {
					return errors.New("heap add stem error - file already exists but should not"), nil
				}

				if node.Info.IsDir() {
					// zlog.
					// 	Trace().
					// 	Str("path", node.Path).
					// 	Msg("heapAddStem / onMatch isDir")

					err = os.Mkdir(destPath, os.ModePerm)
					if err != nil {
						return err, nil
					}
				} else if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
					// zlog.
					// 	Trace().
					// 	Str("path", node.Path).
					// 	Msg("heapAddStem / onMatch isSymlink")

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
					// zlog.
					// 	Trace().
					// 	Str("path", node.Path).
					// 	Msg("heapAddStem / onMatch isFile")

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
							Msg("heapAddStem / onMatch / HeapAddFile error")
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

	// walk the requirements and convert them to dependencies
	requirementsPath := filepath.Join(finalStemPath, "dyd", "requirements")
	dependenciesPath := filepath.Join(finalStemPath, "dyd", "dependencies")
	requirements, err := filepath.Glob(filepath.Join(requirementsPath, "*"))
	if err != nil {
		return err, nil
	}

	for _, requirementPath := range requirements {
		targetFingerprintFile := filepath.Join(requirementPath, "dyd", "fingerprint")
		targetFingerprintBytes, err := ioutil.ReadFile(targetFingerprintFile)
		if err != nil {
			return err, nil
		}
		targetFingerprint := string(targetFingerprintBytes)

		dependencyPath := filepath.Join(dependenciesPath, filepath.Base(requirementPath))
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

	secretsFingerprintPath := filepath.Join(finalStemPath, "dyd", "secrets-fingerprint")

	hasSecrets, err := fileExists(secretsFingerprintPath)
	if err != nil {
		return err, nil
	}

	if hasSecrets {
		secretsPath := filepath.Join(stemPath, "dyd", "secrets")

		err, secretsRef := req.HeapStems.Heap.Secrets().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, secretRef := secretsRef.AddSecret(ctx, secretsPath)
		if err != nil {
			return err, nil
		}

		secretsMountPoint := filepath.Join(finalStemPath, "dyd", "secrets")
		secretsHeapPath := secretRef.BasePath

		relativeLink, err := filepath.Rel(
			filepath.Dir(secretsMountPoint),
			secretsHeapPath,
		)
		if err != nil {
			return err, nil
		}

		err = os.Symlink(relativeLink, secretsMountPoint)
		if err != nil {
			return err, nil
		}

	}


	var setPermissionsShouldCrawl = func (ctx *task.ExecutionContext, node fs2.Walk6Node) (error, bool) {
		isDir := node.Info.IsDir()

		zlog.Trace().
			Str("path", node.VPath).
			Bool("shouldCrawl", isDir).
			Msg("heap add stem - dir ShouldCrawl")

		return nil, isDir
	}	

	var setPermissionsShouldMatch = func(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, bool) {
		isDir := node.Info.IsDir()

		zlog.Trace().
			Str("path", node.VPath).
			Bool("shouldMatch", isDir).
			Msg("heap add stem - dir ShouldMatch")

		return nil, isDir
	}

	var setPermissionsOnMatch = func(ctx *task.ExecutionContext, node fs2.Walk6Node) (error, any) {
		zlog.Trace().
			Str("path", node.VPath).
			Msg("heap add stem - dir OnMatch")

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
			BasePath:    finalStemPath,
			Path:        finalStemPath,
			VPath:       finalStemPath,
			ShouldWalk: setPermissionsShouldCrawl,
			OnPostMatch: fs2.ConditionalWalkAction(setPermissionsOnMatch, setPermissionsShouldMatch),
		},
	)
	if err != nil {
		return err, nil
	}

	stemRef := SafeHeapStemReference{
		BasePath: finalStemPath,
		Stems: req.HeapStems,
	}

	return nil, &stemRef
}

var memoHeapAddStem = task.Memoize(
	heapAddStem,
	func (ctx * task.ExecutionContext, req heapAddStemRequest) (error, any) {
		type Key struct {
			Group string
			Fingerprint string
		}
		var res Key
		var fingerprint string
		var err error

		fingerprint, err = _readFile(
			filepath.Join(req.StemPath, "dyd", "fingerprint"),
		)
		if err != nil {
			return err, res
		}
		
		res = Key{
			Group: "HeapStems.AddStem",
			Fingerprint: fingerprint,
		}

		return nil, res
	},
)

type HeapAddStemRequest struct {
	StemPath string	
}

func (heapStems *SafeHeapStemsReference) AddStem(
	ctx *task.ExecutionContext,
	req HeapAddStemRequest,
) (error, *SafeHeapStemReference) {
	err, heapFiles := heapStems.Heap.Files().Resolve(ctx)
	if err != nil {
		return err, nil
	}

	err, res := memoHeapAddStem(
		ctx,
		heapAddStemRequest{
			HeapStems: heapStems,
			HeapFiles: heapFiles,
			StemPath: req.StemPath,
		},
	)

	return err, res
}