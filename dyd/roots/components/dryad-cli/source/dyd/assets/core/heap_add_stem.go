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

type HeapAddStemRequest struct {
	HeapPath string
	StemPath string	
}

// HeapAddStem takes a stem in a directory, and adds it to the heap.
// the heap path is normalized before adding
func HeapAddStem(ctx *task.ExecutionContext, req HeapAddStemRequest) (error, string) {
	var heapPath = req.HeapPath
	var stemPath = req.StemPath

	// normalize the heap path
	heapPath, err := HeapPath(heapPath)
	if err != nil {
		return err, ""
	}

	gardenFilesPath := filepath.Join(heapPath, "files")
	gardenStemsPath := filepath.Join(heapPath, "stems")

	stemFingerprint, err := _readFile(filepath.Join(stemPath, "dyd", "fingerprint"))
	if err != nil {
		return err, ""
	}

	finalStemPath := filepath.Join(gardenStemsPath, stemFingerprint)

	// check to see if the stem already exists in the garden
	stemExists, err := fileExists(finalStemPath)
	if err != nil {
		return err, ""
	}

	// if stem exists, do nothing
	if stemExists {
		return nil, finalStemPath
	}

	err = os.MkdirAll(finalStemPath, fs.ModePerm)
	if err != nil {
		return err, ""
	}

	// walk the packed root files and copy them into the garden heap
	err = StemWalk(
		ctx,
		StemWalkRequest{
			BasePath: stemPath,
			OnMatch: func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
				zlog.
					Trace().
					Str("path", node.Path).
					Str("vpath", node.VPath).
					Msg("HeapAddStem / onMatch")

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
					// 	Msg("HeapAddStem / onMatch isDir")

					err = os.Mkdir(destPath, os.ModePerm)
					if err != nil {
						return err, nil
					}
				} else if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
					// zlog.
					// 	Trace().
					// 	Str("path", node.Path).
					// 	Msg("HeapAddStem / onMatch isSymlink")

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
					// 	Msg("HeapAddStem / onMatch isFile")

					err, fileFingerprint := HeapAddFile(
						ctx,
						HeapAddFileRequest{
							HeapPath: gardenFilesPath,
							SourcePath: node.Path,
						},
					)
					if err != nil {
						zlog.
							Trace().
							Str("path", node.Path).
							Str("vpath", node.VPath).
							Err(err).
							Msg("HeapAddStem / onMatch / HeapAddFile error")
						return err, nil
					}

					fileHeapPath := filepath.Join(gardenFilesPath, fileFingerprint)

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
		return err, ""
	}

	// walk the requirements and convert them to dependencies
	requirementsPath := filepath.Join(finalStemPath, "dyd", "requirements")
	dependenciesPath := filepath.Join(finalStemPath, "dyd", "dependencies")
	requirements, err := filepath.Glob(filepath.Join(requirementsPath, "*"))
	if err != nil {
		return err, ""
	}

	for _, requirementPath := range requirements {
		targetFingerprintFile := filepath.Join(requirementPath, "dyd", "fingerprint")
		targetFingerprintBytes, err := ioutil.ReadFile(targetFingerprintFile)
		if err != nil {
			return err, ""
		}
		targetFingerprint := string(targetFingerprintBytes)

		dependencyPath := filepath.Join(dependenciesPath, filepath.Base(requirementPath))
		dependencyGardenPath := filepath.Join(gardenStemsPath, targetFingerprint)
		relPath, err := filepath.Rel(dependenciesPath, dependencyGardenPath)
		if err != nil {
			return err, ""
		}

		err = os.Symlink(relPath, dependencyPath)
		if err != nil {
			return err, ""
		}

	}

	secretsFingerprintPath := filepath.Join(finalStemPath, "dyd", "secrets-fingerprint")

	hasSecrets, err := fileExists(secretsFingerprintPath)
	if err != nil {
		return err, ""
	}

	if hasSecrets {
		secretsFingerprint, err := HeapAddSecrets(heapPath, stemPath)
		if err != nil {
			return err, ""
		}

		secretsMountPoint := filepath.Join(finalStemPath, "dyd", "secrets")
		secretsHeapPath := filepath.Join(heapPath, "secrets", secretsFingerprint)

		relativeLink, err := filepath.Rel(
			filepath.Dir(secretsMountPoint),
			secretsHeapPath,
		)
		if err != nil {
			return err, ""
		}

		err = os.Symlink(relativeLink, secretsMountPoint)
		if err != nil {
			return err, ""
		}

	}


	var setPermissionsShouldCrawl = func (ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		isDir := node.Info.IsDir()

		zlog.Trace().
			Str("path", node.VPath).
			Bool("shouldCrawl", isDir).
			Msg("heap add stem - dir ShouldCrawl")

		return nil, isDir
	}	

	var setPermissionsShouldMatch = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		isDir := node.Info.IsDir()

		zlog.Trace().
			Str("path", node.VPath).
			Bool("shouldMatch", isDir).
			Msg("heap add stem - dir ShouldMatch")

		return nil, isDir
	}

	var setPermissionsOnMatch = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
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
	err = fs2.DFSWalk3(
		ctx,
		fs2.Walk5Request{
			Path:        finalStemPath,
			VPath:       finalStemPath,
			BasePath:    finalStemPath,
			ShouldCrawl: setPermissionsShouldCrawl,
			ShouldMatch: setPermissionsShouldMatch,
			OnMatch: setPermissionsOnMatch,
		},
	)
	if err != nil {
		return err, ""
	}

	return nil, finalStemPath
}
