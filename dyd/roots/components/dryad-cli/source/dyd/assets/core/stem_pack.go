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
	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)

type stemPackRequest struct {
	SourceStemPath string
	TargetGarden *SafeGardenReference
}

func stemPack(
	ctx *task.ExecutionContext,
	context BuildContext,
	request stemPackRequest,
) (string, error) {	

	var stemPath = request.SourceStemPath
	var err error

	zlog.
		Debug().
		Str("stemPath", stemPath).
		Str("targetPath", request.TargetGarden.BasePath).
		Msg("StemPack packing stem")

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

		stemPack(
			ctx,
			context,
			stemPackRequest{
				SourceStemPath: dependencyPath,
				TargetGarden: request.TargetGarden,
			},
		)
	}

	err, targetHeap := request.TargetGarden.Heap().Resolve(ctx)
	if err != nil {
		return "", err
	}

	err, targetStems := targetHeap.Stems().Resolve(ctx)
	if err != nil {
		return "", err
	}

	err, packedStem := targetStems.AddStem(
		task.SERIAL_CONTEXT,
		HeapAddStemRequest{
			StemPath: stemPath,
		},
	)
	if err != nil {
		return "", err
	}

	return packedStem.BasePath, nil
}

func finalizeSproutPath(targetGarden *SafeGardenReference, packedStemPath string) (string, error) {
	zlog.
		Debug().
		Str("targetPath", targetGarden.BasePath).
		Str("stemPath", packedStemPath).
		Msg("StemPack / finalizeSproutPath")

	sproutPath := filepath.Join(targetGarden.BasePath, "dyd", "sprouts", "main")
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
	zlog.
		Debug().
		Str("stemPath", request.SourceStemPath).
		Str("targetPath", request.TargetPath).
		Str("format", request.Format).
		Msg("StemPack / stemArchive")

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

		var shouldCrawl = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
			// don't crawl symlinks
			if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
				return nil, false
			}
			return nil, true
		}

		var onMatch = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
			zlog.
				Trace().
				Str("node.Path", node.Path).
				Msg("StemPack/stemArchive/onMatch")

			relativePath, err := filepath.Rel(node.BasePath, node.VPath)
			if err != nil {
				return err, nil
			}

			// remove the name of the base directory
			relativePath = _stripFirstSegment(relativePath)


			if node.Info.IsDir() {
				// create a new dir/file header
				header, err := tar.FileInfoHeader(node.Info, relativePath)
				if err != nil {
					return err, nil
				}
				header.Name = relativePath
				header.Typeflag = tar.TypeDir
	
				err = tarWriter.WriteHeader(header)
				if err != nil {
					return err, nil
				}
	
			} else if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
				// if it's a symlink, read the link target
				linkPath, err := os.Readlink(node.Path)
				if err != nil {
					return err, nil
				}
	
				// create a new dir/file header
				header, err := tar.FileInfoHeader(node.Info, relativePath)
				if err != nil {
					return err, nil
				}
				header.Name = relativePath
				header.Typeflag = tar.TypeSymlink
				header.Linkname = linkPath
	
				err = tarWriter.WriteHeader(header)
				if err != nil {
					return err, nil
				}
			} else if node.Info.Mode().IsRegular() {
				_, hashString, err := fileHash(node.Path)
				if err != nil {
					return err, nil
				}

				existingPath, hasExistingPath := packMap[hashString]

				if hasExistingPath {
					// create a new hard link header
					header, err := tar.FileInfoHeader(node.Info, relativePath)
					if err != nil {
						return err, nil
					}
					header.Name = relativePath
					header.Typeflag = tar.TypeLink
					header.Linkname = existingPath

					err = tarWriter.WriteHeader(header)
					if err != nil {
						return err, nil
					}

				} else {
				// create a new dir/file header
					header, err := tar.FileInfoHeader(node.Info, relativePath)
					if err != nil {
						return err, nil
					}
					header.Name = relativePath
					header.Typeflag = tar.TypeReg

					err = tarWriter.WriteHeader(header)
					if err != nil {
						return err, nil
					}

					// add path to the packMap
					packMap[hashString] = relativePath

					file, err := os.Open(node.Path)
					if err != nil {
						return err, nil
					}
					defer file.Close()

					_, err = io.Copy(tarWriter, file)
					if err != nil {
						return err, nil
					}
				}				
			}

			return nil, nil
		}

		// NOTE: packing needs to be serial for now
		err, _ = fs2.BFSWalk3(
			task.SERIAL_CONTEXT,
			fs2.Walk5Request{
				BasePath:    request.TargetPath,
				Path:        request.TargetPath,
				VPath:       request.TargetPath,
				OnMatch:     onMatch,
				ShouldCrawl: shouldCrawl,
			},
		)
		if err != nil {
			return "", err
		}

		// clear the archive directory
		err, _ = fs2.RemoveAll(task.SERIAL_CONTEXT, request.TargetPath)
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

type StemPackRequest struct {
	SourceGarden *SafeGardenReference
	SourceStemPath string
	TargetPath string
	Format string
}

func StemPack(
	ctx *task.ExecutionContext,
	request StemPackRequest,
) (string, error) {
	zlog.
		Debug().
		Str("stemPath", request.SourceStemPath).
		Str("targetPath", request.TargetPath).
		Str("format", request.Format).
		Msg("StemPack")

	var buildContext BuildContext = BuildContext{
		Fingerprints: map[string]string{},
	}

	err := os.MkdirAll(request.TargetPath, os.ModePerm)
	if err != nil {
		return "", err
	}	
	
	var unsafeTargetGardenRef = Garden(request.TargetPath)
	// var targetGardenRef *SafeGardenReference

	err, safeTargetGardenRef := unsafeTargetGardenRef.Create(ctx)
	if err != nil {
		return "", err
	}

	packedStemPath, err := stemPack(
		ctx,
		buildContext,
		stemPackRequest{
			TargetGarden: safeTargetGardenRef,
			SourceStemPath: request.SourceStemPath,
		},
	)
	if err != nil {
		return "", err
	}

	_, err = finalizeSproutPath(safeTargetGardenRef, packedStemPath)
	if err != nil {
		return "", err
	}

	path, err := stemArchive(request)

	return path, err
}
