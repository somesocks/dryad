package core

import (
	fs2 "dryad/filesystem"
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

// HeapAddStem takes a stem in a directory, and adds it to the heap.
// the heap path is normalized before adding
func HeapAddStem(heapPath string, stemPath string) (string, error) {
	// // fmt.Println("[trace] HeapAddStem", heapPath, stemPath)

	// normalize the heap path
	heapPath, err := HeapPath(heapPath)
	if err != nil {
		return "", err
	}

	gardenFilesPath := filepath.Join(heapPath, "files")
	gardenStemsPath := filepath.Join(heapPath, "stems")

	stemFingerprint, err := _readFile(filepath.Join(stemPath, "dyd", "fingerprint"))
	if err != nil {
		return "", err
	}

	finalStemPath := filepath.Join(gardenStemsPath, stemFingerprint)

	// check to see if the stem already exists in the garden
	stemExists, err := fileExists(finalStemPath)
	if err != nil {
		return "", err
	}

	// if stem exists, do nothing
	if stemExists {
		return finalStemPath, nil
	}

	err = os.MkdirAll(finalStemPath, fs.ModePerm)
	if err != nil {
		return "", err
	}

	// walk the packed root files and copy them into the garden heap
	err = StemWalk(
		StemWalkRequest{
			BasePath: stemPath,
			OnMatch: func(context fs2.Walk4Context) error {
				zlog.
					Trace().
					Str("path", context.Path).
					Str("vpath", context.VPath).
					Msg("HeapAddStem / onMatch")

				relPath, err := filepath.Rel(context.BasePath, context.VPath)
				if err != nil {
					return err
				}

				destPath := filepath.Join(finalStemPath, relPath)

				// if the file already exists, we hit it on a previous pass through a symlink
				destExists, err := fileExists(destPath)
				if err != nil {
					return err
				}
				if destExists {
					return errors.New("heap add stem error - file already exists but should not")
				}

				if context.Info.IsDir() {
					// zlog.
					// 	Trace().
					// 	Str("path", context.Path).
					// 	Msg("HeapAddStem / onMatch isDir")

					err = os.Mkdir(destPath, os.ModePerm)
					if err != nil {
						return err
					}
				} else if context.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
					// zlog.
					// 	Trace().
					// 	Str("path", context.Path).
					// 	Msg("HeapAddStem / onMatch isSymlink")

					linkTarget, err := os.Readlink(context.Path)
					if err != nil {
						return err
					}
										
					absLinkTarget := linkTarget
					if !filepath.IsAbs(absLinkTarget) {
						absLinkTarget = filepath.Join(filepath.Dir(context.VPath), linkTarget)
					} 
						
					isInternalLink, err := fileIsDescendant(absLinkTarget, context.BasePath)
			
					if isInternalLink {
						err = os.Symlink(linkTarget, destPath)
						if err != nil {
							return err
						}
					} 
				} else {
					// zlog.
					// 	Trace().
					// 	Str("path", context.Path).
					// 	Msg("HeapAddStem / onMatch isFile")

					fileFingerprint, err := HeapAddFile(gardenFilesPath, context.Path)
					if err != nil {
						return err
					}

					fileHeapPath := filepath.Join(gardenFilesPath, fileFingerprint)

					err = os.Link(fileHeapPath, destPath)
					if err != nil {
						return err
					}
				}

				return nil
			},
		},
	)
	if err != nil {
		return "", err
	}

	// walk the requirements and convert them to dependencies
	requirementsPath := filepath.Join(finalStemPath, "dyd", "requirements")
	dependenciesPath := filepath.Join(finalStemPath, "dyd", "dependencies")
	requirements, err := filepath.Glob(filepath.Join(requirementsPath, "*"))
	if err != nil {
		return "", err
	}

	for _, requirementPath := range requirements {
		targetFingerprintFile := filepath.Join(requirementPath, "dyd", "fingerprint")
		targetFingerprintBytes, err := ioutil.ReadFile(targetFingerprintFile)
		if err != nil {
			return "", err
		}
		targetFingerprint := string(targetFingerprintBytes)

		dependencyPath := filepath.Join(dependenciesPath, filepath.Base(requirementPath))
		dependencyGardenPath := filepath.Join(gardenStemsPath, targetFingerprint)
		relPath, err := filepath.Rel(dependenciesPath, dependencyGardenPath)
		if err != nil {
			return "", err
		}

		err = os.Symlink(relPath, dependencyPath)
		if err != nil {
			return "", err
		}

	}

	secretsFingerprintPath := filepath.Join(finalStemPath, "dyd", "secrets-fingerprint")

	hasSecrets, err := fileExists(secretsFingerprintPath)
	if err != nil {
		return "", err
	}

	if hasSecrets {
		secretsFingerprint, err := HeapAddSecrets(heapPath, stemPath)
		if err != nil {
			return "", err
		}

		secretsMountPoint := filepath.Join(finalStemPath, "dyd", "secrets")
		secretsHeapPath := filepath.Join(heapPath, "secrets", secretsFingerprint)

		relativeLink, err := filepath.Rel(
			filepath.Dir(secretsMountPoint),
			secretsHeapPath,
		)
		if err != nil {
			return "", err
		}

		err = os.Symlink(relativeLink, secretsMountPoint)
		if err != nil {
			return "", err
		}

	}

	// now that all files are added, sweep through in a second pass and make directories read-only
	err = fs2.DFSWalk2(fs2.Walk4Request{
		Path:        finalStemPath,
		VPath:       finalStemPath,
		BasePath:    finalStemPath,
		ShouldCrawl: func(context fs2.Walk4Context) (bool, error) {
			isDir := context.Info.IsDir()

			zlog.Trace().
				Str("path", context.VPath).
				Bool("shouldCrawl", isDir).
				Msg("heap add stem - dir ShouldCrawl")

			return isDir, nil
		},
		ShouldMatch: func(context fs2.Walk4Context) (bool, error) {
			isDir := context.Info.IsDir()

			zlog.Trace().
				Str("path", context.VPath).
				Bool("shouldMatch", isDir).
				Msg("heap add stem - dir ShouldMatch")

			return isDir, nil
		},
		OnMatch: func(context fs2.Walk4Context) error {
			zlog.Trace().
				Str("path", context.VPath).
				Msg("heap add stem - dir OnMatch")
	
			dirPerms := context.Info.Mode().Perm()

			// if permissions are already set correctly, do nothing
			if dirPerms == 0o511 {
				return nil
			}

			dir, err := os.Open(context.Path)
			if err != nil {
				return err
			}
			defer dir.Close()

			// heap files should be set to R-X--X--X
			err = dir.Chmod(0o511)
			if err != nil {
				return err
			}

			return nil
		},
	})
	if err != nil {
		return "", err
	}

	return finalStemPath, nil
}
