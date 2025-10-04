package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"io/fs"

)


type rootCopyRequest struct {
	Source *SafeRootReference
	Dest *UnsafeRootReference
	Unpin bool
}

type typeRootCopy = func (*task.ExecutionContext, rootCopyRequest) (error, *SafeRootReference)

var rootCopy typeRootCopy = func () typeRootCopy {

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
			"|(dyd/secrets)" +
			"|(dyd/secrets/.*)" +
			"|(dyd/traits)" +
			"|(dyd/traits/.*)" +
			")$",
	)

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
			"|(dyd/traits)" +
			"|(dyd/traits/.*)" +
			")$",
	)
		

	// don't crawl symlinks
	var shouldWalk = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		var relPath, relErr = filepath.Rel(node.BasePath, node.Path)
		if relErr != nil {
			return relErr, false
		}

		if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
			return nil, false
		}

		return nil, _ROOT_COPY_CRAWL_INCLUDE_REGEXP.Match([]byte(relPath))
	}


	var shouldMatch = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		var relPath, relErr = filepath.Rel(node.BasePath, node.Path)
		if relErr != nil {
			return relErr, false
		}

		res := _ROOT_COPY_MATCH_INCLUDE_REGEXP.Match([]byte(relPath))

		return nil, res
	}

	var copyDir = func (ctx *task.ExecutionContext, path string, mode fs.FileMode) (error) {
		// for a directory, make a new dir
		var err = os.MkdirAll(path, mode)
		return err
	}

	var copySymlink = func (
		ctx *task.ExecutionContext,
		basePath string,
		sourcePath string, 
		destPath string,
		unpinMode bool,
	) (error) {
		var linkTarget string
		var absLinkTarget string
		var newLinkTarget string
		var isInternalLink bool
		var err error


		linkTarget, err = os.Readlink(sourcePath)
		if err != nil {
			return err
		}

		if !filepath.IsAbs(linkTarget) {
			absLinkTarget = filepath.Join(
				filepath.Dir(sourcePath),
				linkTarget,
			)
		} else {
			absLinkTarget = linkTarget
		}

		isInternalLink, err = fileIsDescendant(absLinkTarget, basePath)

		if isInternalLink {
			newLinkTarget = linkTarget
		} else {
			if unpinMode {
				absLinkTarget = filepath.Join(
					filepath.Dir(destPath),
					linkTarget,
				)
				targetExists, err := fileExists(absLinkTarget)
				if err != nil {
					return err
				}
				if !targetExists {
					absLinkTarget = filepath.Join(
						filepath.Dir(sourcePath),
						linkTarget,
					)
				}
			}

			newLinkTarget, err = filepath.Rel(
				filepath.Dir(destPath),
				absLinkTarget,
			)
			if err != nil {
				return err
			}
		}
	
		err = os.Symlink(newLinkTarget, destPath)
		return err
	}

	var copyFile = func (ctx *task.ExecutionContext, sourcePath string, sourceMode fs.FileMode, destPath string) error {	
		// for a file, copy contents

		srcFile, err := os.Open(sourcePath)
		if err != nil {
			return err
		}
		defer srcFile.Close()

		var destFile *os.File
		destFile, err = os.Create(destPath)
		if err != nil {
			return err
		}
		defer destFile.Close()

		_, err = destFile.ReadFrom(srcFile)
		if err != nil {
			return err
		}

		err = destFile.Chmod(sourceMode)
		if err != nil {
			return err
		}

		return nil
	}

	var rootCopy = func (ctx *task.ExecutionContext, req rootCopyRequest) (error, *SafeRootReference) {
		var sourcePath string = req.Source.BasePath
		var destPath string = req.Dest.BasePath
		var err error
	
		// check that source and destination are within the same garden
		if req.Source.Roots.Garden.BasePath != req.Dest.Roots.Garden.BasePath {
			return fmt.Errorf("source and destination roots are not in same garden"), nil
		}
	
		onMatch := func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
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
				err = copyDir(ctx, targetDestPath, node.Info.Mode())
				return err, nil
			} else if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
				err = copySymlink(
					ctx,
					node.BasePath,
					node.Path,
					targetDestPath,
					req.Unpin,
				)
				return err, nil
			} else {
				err = copyFile(
					ctx,
					node.Path,
					node.Info.Mode(),
					targetDestPath,
				)
				return err, nil
			}
		}
	
		onMatch = dydfs.ConditionalWalkAction(onMatch, shouldMatch)

		err, _ = dydfs.Walk6(
			ctx,
			dydfs.Walk6Request{
				BasePath:     sourcePath,
				Path:     sourcePath,
				VPath:     sourcePath,
				ShouldWalk: shouldWalk,
				OnPreMatch:      onMatch,
			},
		)
		if err != nil {
			return err, nil
		}

		err, sourceRequirements := req.Source.Requirements().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		var newRoot SafeRootReference
		err, newRoot = req.Dest.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, newRequirements := newRoot.Requirements().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		sourceRequirements.Walk(
			ctx,
			RootRequirementsWalkRequest{
				OnMatch: func (
					ctx *task.ExecutionContext,
					requirement *SafeRootRequirementReference,
				) (error, any) {
					err, _ := requirement.Copy(ctx, RootRequirementCopyRequest{
						DestRequirements: newRequirements,
						Unpin: req.Unpin,
					})
					return err, nil
				},
			},
		)
	
		return nil, &newRoot
	}
	
	return rootCopy
}();


type RootCopyRequest struct {
	Dest *UnsafeRootReference
	Unpin bool
}

func (root *SafeRootReference) Copy(ctx *task.ExecutionContext, req RootCopyRequest) (error, *SafeRootReference) {
	err, res := rootCopy(
		ctx,
		rootCopyRequest{
			Source: root,
			Dest: req.Dest,
			Unpin: req.Unpin,
		},
	)
	return err, res
}