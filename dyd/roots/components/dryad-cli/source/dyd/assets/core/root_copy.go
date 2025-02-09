package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var _ROOT_COPY_CRAWL_INCLUDE_REGEXP = regexp.MustCompile(
	"^(" +
		"(\\.)" +
		"|(dyd)" +
		"|(dyd/assets)" +
		"|(dyd/assets/.*)" +
		"|(dyd/commands)" +
		"|(dyd/commands/.*)" +
		"|(dyd/docs)" +
		"|(dyd/docs/.*)" +
		"|(dyd/requirements)" +
		"|(dyd/requirements/.*)" +
		"|(dyd/secrets)" +
		"|(dyd/secrets/.*)" +
		"|(dyd/traits)" +
		"|(dyd/traits/.*)" +
		")$",
)

var _ROOT_COPY_CRAWL_EXCLUDE_REGEXP = regexp.MustCompile(`^$`)

var _ROOT_COPY_MATCH_INCLUDE_REGEXP = regexp.MustCompile(
	"^(" +
		"(\\.)" +
		"|(dyd)" +
		"|(dyd/assets)" +
		"|(dyd/assets/.*)" +
		"|(dyd/commands)" +
		"|(dyd/commands/.*)" +
		"|(dyd/docs)" +
		"|(dyd/docs/.*)" +
		"|(dyd/secrets)" +
		"|(dyd/secrets/.*)" +
		"|(dyd/fingerprint)" +
		"|(dyd/type)" +
		"|(dyd/root)" +
		"|(dyd/secrets-fingerprint)" +
		"|(dyd/requirements)" +
		"|(dyd/requirements/.*)" +
		"|(dyd/traits)" +
		"|(dyd/traits/.*)" +
		")$",
)

var _ROOT_COPY_MATCH_EXCLUDE_REGEXP = regexp.MustCompile(`^$`)

type RootCopyRequest struct {
	Garden *SafeGardenReference
	SourcePath string
	DestPath string
}

func RootCopy(ctx *task.ExecutionContext, req RootCopyRequest) (error, any) {
	var sourcePath string = req.SourcePath
	var destPath string = req.DestPath

	// normalize the source path
	sourcePath, err := RootPath(sourcePath, "")
	if err != nil {
		return err, nil
	}

	// normalize the destination path
	destPath, err = filepath.Abs(destPath)
	if err != nil {
		return err, nil
	}

	rootsPath, err := RootsPath(req.Garden.BasePath)
	if err != nil {
		return err, nil
	}

	isWithinRoots, err := fileIsDescendant(destPath, rootsPath)
	if err != nil {
		return err, nil
	}

	if !isWithinRoots {
		return fmt.Errorf("destination path %s is outside of roots", destPath), nil
	}

	// gardenPath, err := GardenPath(sourcePath)
	// if err != nil {
	// 	return err
	// }

	// don't crawl symlinks
	shouldCrawl := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		var relPath, relErr = filepath.Rel(node.BasePath, node.Path)
		if relErr != nil {
			return relErr, false
		}

		if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil, false
		}

		return nil, _ROOT_COPY_CRAWL_INCLUDE_REGEXP.Match([]byte(relPath))
	}

	shouldMatch := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		var relPath, relErr = filepath.Rel(node.BasePath, node.Path)
		if relErr != nil {
			return relErr, false
		}

		res := _ROOT_COPY_MATCH_INCLUDE_REGEXP.Match([]byte(relPath))

		return nil, res
	}

	onMatch := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
		var relPath, relErr = filepath.Rel(node.BasePath, node.Path)
		if relErr != nil {
			return relErr, nil
		}

		targetDestPath := filepath.Join(destPath, relPath)
		targetDestExists, err := fileExists(targetDestPath)
		if err != nil {
			return err, nil
		} else if targetDestExists {
			return fmt.Errorf("error: copy destination %s already exists", targetDestPath), nil
		}

		if node.Info.IsDir() {
			// for a directory, make a new dir

			err = os.MkdirAll(targetDestPath, node.Info.Mode())
			return err, nil

		} else if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
			// for a symlink, make a new link resolving to the target

			linkPath, err := os.Readlink(node.Path)
			if err != nil {
				return err, nil
			}

			// convert relative links to an absolute path
			if !filepath.IsAbs(linkPath) {
				linkPath = filepath.Join(
					filepath.Dir(node.Path),
					linkPath,
				)
			}

			// fmt.Println("[debug] root copy symlink linkpath", linkPath)

			linkRelPath, err := filepath.Rel(filepath.Dir(targetDestPath), linkPath)
			if err != nil {
				return err, nil
			}

			err = os.Symlink(linkRelPath, targetDestPath)
			return err, nil

		} else {

			// for a file, copy contents

			srcFile, err := os.Open(node.Path)
			if err != nil {
				return err, nil
			}
			defer srcFile.Close()

			var destFile *os.File
			destFile, err = os.Create(targetDestPath)
			if err != nil {
				return err, nil
			}
			defer destFile.Close()

			_, err = destFile.ReadFrom(srcFile)
			if err != nil {
				return err, nil
			}

			err = destFile.Chmod(node.Info.Mode())
			if err != nil {
				return err, nil
			}

			return nil, nil
		}
	}

	err, _ = fs2.BFSWalk3(
		ctx,
		fs2.Walk5Request{
			BasePath:     sourcePath,
			Path:     sourcePath,
			VPath:     sourcePath,
			ShouldCrawl: shouldCrawl,
			ShouldMatch: shouldMatch,
			OnMatch:      onMatch,
		},
	)
	return err, nil
}
