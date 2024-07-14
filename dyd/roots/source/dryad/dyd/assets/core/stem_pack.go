package core

import (
	"archive/tar"
	"compress/gzip"
	// "errors"
	"io"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"fmt"
	"strings"

	fs2 "dryad/filesystem"
	zlog "github.com/rs/zerolog/log"
)

type StemPackRequest struct {
	StemPath string
	TargetPath string
	Format string
}

func stemPack(context BuildContext, request StemPackRequest) (string, error) {
	var stemPath = request.StemPath
	var targetPath = request.TargetPath
	var err error

	// convert relative stem path to absolute
	stemPath, err = filepath.Abs(stemPath) 
	if err != nil {
		return "", err
	}

	// resolve the dir to the root of the stem
	stemPath, err = StemPath(stemPath)
	if err != nil {
		return "", err
	}

	// convert relative target to absolute
	targetPath, err = filepath.Abs(targetPath) 
	if err != nil {
		return "", err
	}	

	stemFingerprint, err := _readFile(filepath.Join(stemPath, "dyd", "fingerprint"))
	if err != nil {
		return "", err
	}

	_, contextHasFingerprint := context.Fingerprints[stemPath]

	if contextHasFingerprint {
		return stemPath, nil
	}

	context.Fingerprints[stemPath] = stemFingerprint	

	// walk through the dependencies, and add them to the archive
	dependenciesPath := filepath.Join(stemPath, "dyd", "dependencies")

	dependencies, err := filepath.Glob(filepath.Join(dependenciesPath, "*"))
	if err != nil {
		return "", err
	}

	for _, dependencyPath := range dependencies {
		dependencyPath, err = filepath.EvalSymlinks(dependencyPath)
		if err != nil {
			return "", err
		}

		stemPack(context, StemPackRequest{
			StemPath: dependencyPath,
			TargetPath: targetPath,
			Format: request.Format,
		})
	}

	var packedStemPath string
	packedStemPath, err = HeapAddStem(targetPath, stemPath)
	if err != nil {
		return "", err
	}

	sproutPath := filepath.Join(targetPath, "dyd", "sprouts", "main")
	sproutParent := filepath.Dir(sproutPath)
	sproutHeapPath := packedStemPath
	relSproutLink, err := filepath.Rel(
		sproutParent,
		sproutHeapPath,
	)
	if err != nil {
		return "", err
	}

	// fmt.Println("[debug] building sprout parent")
	err = fs2.MkDir(sproutParent, fs.ModePerm)
	if err != nil {
		return "", err
	}

	// fmt.Println("[debug] setting write permission on sprout parent")
	err = os.Chmod(sproutParent, 0o711)
	if err != nil {
		return "", err
	}

	tmpSproutPath := sproutPath + ".tmp"
	// fmt.Println("[debug] adding temporary sprout link")
	err = os.Symlink(relSproutLink, tmpSproutPath)
	if err != nil {
		return "", err
	}

	// fmt.Println("[debug] renaming sprout link", sproutPath)
	err = os.Rename(tmpSproutPath, sproutPath)
	if err != nil {
		return "", err
	}

	// fmt.Println("[debug] setting read permissions on sprout parent")
	err = os.Chmod(sproutParent, 0o511)
	if err != nil {
		return "", err
	}

	return sproutPath, nil
}

func _stripFirstSegment(path string) string {
	// Split the path into segments
	segments := strings.Split(path, string(filepath.Separator))

	// Check if there are more than one segment
	if len(segments) > 1 {
		// Join the segments back together, omitting the first one
		return filepath.Join(segments[1:]...)
	}

	// If there is only one segment, return an empty string
	return ""
}


func stemArchive(request StemPackRequest) (string, error) {
	var err error

	switch request.Format {
	case "dir":
		return request.TargetPath, nil
	case "tar", "tar.gz":
		var outputPath string
		var outputWriter *os.File
		var tarWriter *tar.Writer
		var packMap = make(map[string]string)

		if request.Format == "tar.gz" {
			outputPath = path.Clean(request.TargetPath) + ".tar.gz"
			outputWriter, err = os.Create(outputPath)
			if err != nil {
				return "", err
			}
			defer outputWriter.Close()
	
			var gzw = gzip.NewWriter(outputWriter)
			defer gzw.Close()

			tarWriter = tar.NewWriter(gzw)
			defer tarWriter.Close()
		} else {
			outputPath = path.Clean(request.TargetPath) + ".tar"
			outputWriter, err = os.Create(outputPath)
			if err != nil {
				return "", err
			}
			defer outputWriter.Close()
	
			tarWriter = tar.NewWriter(outputWriter)
			defer tarWriter.Close()
		}

		var shouldCrawl = func(context fs2.Walk4Context) (bool, error) {
			// don't crawl symlinks
			if context.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
				return false, nil
			}
			return true, nil
		}

		var onMatch = func(context fs2.Walk4Context) error {
			zlog.
				Trace().
				Str("context.Path", context.Path).
				Msg("StemPack/stemArchive/onMatch")

			relativePath, err := filepath.Rel(context.BasePath, context.VPath)
			if err != nil {
				return err
			}

			// remove the name of the base directory
			relativePath = _stripFirstSegment(relativePath)


			if context.Info.IsDir() {
				// create a new dir/file header
				header, err := tar.FileInfoHeader(context.Info, relativePath)
				if err != nil {
					return err
				}
				header.Name = relativePath
				header.Typeflag = tar.TypeDir
	
				err = tarWriter.WriteHeader(header)
				if err != nil {
					return err
				}
	
			} else if context.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
				// if it's a symlink, read the link target
				linkPath, err := os.Readlink(context.Path)
				if err != nil {
					return err
				}
	
				// create a new dir/file header
				header, err := tar.FileInfoHeader(context.Info, relativePath)
				if err != nil {
					return err
				}
				header.Name = relativePath
				header.Typeflag = tar.TypeSymlink
				header.Linkname = linkPath
	
				err = tarWriter.WriteHeader(header)
				if err != nil {
					return err
				}
			} else if context.Info.Mode().IsRegular() {
				_, hashString, err := fileHash(context.Path)
				if err != nil {
					return err
				}

				existingPath, hasExistingPath := packMap[hashString]

				if hasExistingPath {
					// create a new hard link header
					header, err := tar.FileInfoHeader(context.Info, relativePath)
					if err != nil {
						return err
					}
					header.Name = relativePath
					header.Typeflag = tar.TypeLink
					header.Linkname = existingPath

					err = tarWriter.WriteHeader(header)
					if err != nil {
						return err
					}

				} else {
				// create a new dir/file header
					header, err := tar.FileInfoHeader(context.Info, relativePath)
					if err != nil {
						return err
					}
					header.Name = relativePath
					header.Typeflag = tar.TypeReg

					err = tarWriter.WriteHeader(header)
					if err != nil {
						return err
					}

					// add path to the packMap
					packMap[hashString] = relativePath

					file, err := os.Open(context.Path)
					if err != nil {
						return err
					}
					defer file.Close()

					_, err = io.Copy(tarWriter, file)
					if err != nil {
						return err
					}
				}				
			}

			return nil
		}

		err = fs2.BFSWalk2(fs2.Walk4Request{
			BasePath:    request.TargetPath,
			Path:        request.TargetPath,
			VPath:       request.TargetPath,
			OnMatch:     onMatch,
			ShouldCrawl: shouldCrawl,
		})
		if err != nil {
			return "", err
		}

		// clear the archive directory
		err = fs2.RemoveAll(request.TargetPath)
		if err != nil {
			return "", err
		}

		return outputPath, err
	// case "tar.gz":
	// 	break;
	default:
		return "", fmt.Errorf("unrecognized pack format %s", request.Format)
	}


}

func StemPack(request StemPackRequest) (string, error) {
	var buildContext BuildContext = BuildContext{
		Fingerprints: map[string]string{},
	}

	err := os.MkdirAll(request.TargetPath, os.ModePerm)
	if err != nil {
		return "", err
	}	
	
	err = GardenCreate(request.TargetPath)
	if err != nil {
		return "", err
	}

	_, err = stemPack(buildContext, request)
	if err != nil {
		return "", err
	}

	path, err := stemArchive(request)

	return path, err
}
