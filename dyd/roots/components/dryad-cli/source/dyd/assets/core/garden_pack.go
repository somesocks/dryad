package core

import (
	"archive/tar"
	"compress/gzip"
	fs2 "dryad/filesystem"
	"dryad/task"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

type GardenPackRequest struct {
	Garden *SafeGardenReference
	TargetPath string
	IncludeRoots bool
	IncludeHeap bool
	IncludeContexts bool
	IncludeSprouts bool
	IncludeShed bool
}

func GardenPack(ctx *task.ExecutionContext, req GardenPackRequest) (error, string) {
	var gardenPath = req.Garden.BasePath
	var targetPath = req.TargetPath
	var err error
	
	// convert relative target to absolute
	if !filepath.IsAbs(targetPath) {
		wd, err := os.Getwd()
		if err != nil {
			return err, ""
		}
		targetPath = filepath.Join(wd, targetPath)
	}

	// build archive name
	targetInfo, err := os.Stat(targetPath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err, ""
	} else if targetInfo.IsDir() {
		baseName := filepath.Base(gardenPath + ".tar.gz")
		targetPath = filepath.Join(targetPath, baseName)
	}

	file, err := os.Create(targetPath)
	if err != nil {
		return err, ""
	}
	defer file.Close()

	var gzw = gzip.NewWriter(file)
	defer gzw.Close()

	var tw = tar.NewWriter(gzw)
	defer tw.Close()

	var packMutex sync.Mutex
	var packMap = make(map[string]bool)

	var packEntry = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
		zlog.Trace().
			Str("path", node.Path).
			Msg("GardenPack/packEntry")

		var relativePath string
		var err error

		relativePath, err = filepath.Rel(gardenPath, node.Path)
		if err != nil {
			return err, nil
		}

		// acquire the packing mutex before writing to the tar,
		// or using the pack map
		packMutex.Lock()
		defer packMutex.Unlock()

		// don't pack a file that's already been packed
		if _, ok := packMap[relativePath]; ok {
			return nil, nil
		}

		if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink {
			linkPath, err := os.Readlink(node.Path)
			if err != nil {
				return err, nil
			}

			header, err := tar.FileInfoHeader(node.Info, relativePath)
			if err != nil {
				return err, nil
			}
			header.Name = relativePath
			header.Typeflag = tar.TypeSymlink
			header.Linkname = linkPath

			err = tw.WriteHeader(header)
			if err != nil {
				return err, nil
			}

			// add path to the packMap
			packMap[relativePath] = true

		} else if node.Info.IsDir() {
			// create a new dir/file header
			header, err := tar.FileInfoHeader(node.Info, relativePath)
			if err != nil {
				return err, nil
			}
			header.Name = relativePath
			header.Typeflag = tar.TypeDir

			err = tw.WriteHeader(header)
			if err != nil {
				return err, nil
			}

			// add path to the packMap
			packMap[relativePath] = true
		} else if node.Info.Mode().IsRegular() {
			// create a new dir/file header
			header, err := tar.FileInfoHeader(node.Info, relativePath)
			if err != nil {
				return err, nil
			}
			header.Name = relativePath
			header.Typeflag = tar.TypeReg

			err = tw.WriteHeader(header)
			if err != nil {
				return err, nil
			}

			// add path to the packMap
			packMap[relativePath] = true

			file, err := os.Open(node.Path)
			if err != nil {
				return err, nil
			}
			defer file.Close()

			_, err = io.Copy(tw, file)
			if err != nil {
				return err, nil
			}
		}

		return nil, nil
	}

	packDirShouldCrawl := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		var relativePath string
		var err error

		relativePath, err = filepath.Rel(node.BasePath, node.Path)
		if err != nil {
			return err, false
		}

		var isBase = (relativePath == ".")
		var isDyd = (relativePath == "dyd")
		var isInRoots = (relativePath == "dyd/roots") ||
			strings.HasPrefix(relativePath, "dyd/roots/") 
		var isInHeap = (relativePath == "dyd/heap") ||
			strings.HasPrefix(relativePath, "dyd/heap/")
		var isInContexts = (relativePath == "dyd/heap/contexts") ||
			strings.HasPrefix(relativePath, "dyd/heap/contexts/")
		var isInShed = (relativePath == "dyd/shed") ||
			strings.HasPrefix(relativePath, "dyd/shed/") 
		var isInSprouts = (relativePath == "dyd/sprouts") ||
			strings.HasPrefix(relativePath, "dyd/sprouts/") 

		var shouldCrawl = 
			isBase ||
			isDyd ||
			(isInRoots && req.IncludeRoots) ||
			(isInHeap && req.IncludeHeap) ||
			(isInContexts && req.IncludeContexts) ||
			(isInShed && req.IncludeShed) ||
			(isInSprouts && req.IncludeSprouts)

		return nil, shouldCrawl
	}

	packDirShouldMatch := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		var relativePath string
		var err error

		relativePath, err = filepath.Rel(node.BasePath, node.Path)
		if err != nil {
			return err, false
		}

		var isBase = (relativePath == ".")
		var isDyd = (relativePath == "dyd")
		var isInRoots = (relativePath == "dyd/roots") ||
			strings.HasPrefix(relativePath, "dyd/roots/") 
		var isInHeap = (relativePath == "dyd/heap") ||
			strings.HasPrefix(relativePath, "dyd/heap/") 
		var isInContexts = (relativePath == "dyd/heap/contexts") ||
			strings.HasPrefix(relativePath, "dyd/heap/contexts/")
		var isInShed = (relativePath == "dyd/shed") ||
			strings.HasPrefix(relativePath, "dyd/shed/") 
		var isInSprouts = (relativePath == "dyd/sprouts") ||
			strings.HasPrefix(relativePath, "dyd/sprouts/") 

		var shouldMatch = 
			isBase ||
			isDyd ||
			(isInRoots && req.IncludeRoots) ||
			(isInHeap && req.IncludeHeap) ||
			(isInContexts && req.IncludeContexts) ||
			(isInShed && req.IncludeShed) ||
			(isInSprouts && req.IncludeSprouts)
		shouldMatch = shouldMatch && node.Info.IsDir()

		return nil, shouldMatch
	}

	err, _ = fs2.BFSWalk3(
		ctx,
		fs2.Walk5Request{
			BasePath: gardenPath,
			Path: gardenPath,
			VPath: gardenPath,
			ShouldCrawl: packDirShouldCrawl,
			ShouldMatch: packDirShouldMatch,
			OnMatch:     packEntry,
		},
	)
	if err != nil {
		return err, ""
	}



	packFilesShouldCrawl := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		var relativePath string
		var err error

		relativePath, err = filepath.Rel(node.BasePath, node.Path)
		if err != nil {
			return err, false
		}

		var isBase = (relativePath == ".")
		var isDyd = (relativePath == "dyd")
		var isInRoots = (relativePath == "dyd/roots") ||
			strings.HasPrefix(relativePath, "dyd/roots/") 
		var isInHeap = (relativePath == "dyd/heap") ||
			strings.HasPrefix(relativePath, "dyd/heap/") 
		var isInContexts = (relativePath == "dyd/heap/contexts") ||
			strings.HasPrefix(relativePath, "dyd/heap/contexts/")
		var isInShed = (relativePath == "dyd/shed") ||
			strings.HasPrefix(relativePath, "dyd/shed/") 
		var isInSprouts = (relativePath == "dyd/sprouts") ||
			strings.HasPrefix(relativePath, "dyd/sprouts/") 

		var shouldCrawl = 
			isBase ||
			isDyd ||
			(isInRoots && req.IncludeRoots) ||
			(isInHeap && req.IncludeHeap) ||
			(isInContexts && req.IncludeContexts) ||
			(isInShed && req.IncludeShed) ||
			(isInSprouts && req.IncludeSprouts)

		return nil, shouldCrawl
	}

	packFilesShouldMatch := func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		var relativePath string
		var err error

		relativePath, err = filepath.Rel(node.BasePath, node.Path)
		if err != nil {
			return err, false
		}

		var isBase = (relativePath == ".")
		var isDyd = (relativePath == "dyd")
		var isTypeFile = (relativePath == "dyd/type")
		var isInRoots = (relativePath == "dyd/roots") ||
			strings.HasPrefix(relativePath, "dyd/roots/") 
		var isInHeap = (relativePath == "dyd/heap") ||
			strings.HasPrefix(relativePath, "dyd/heap/") 
		var isInContexts = (relativePath == "dyd/heap/contexts") ||
			strings.HasPrefix(relativePath, "dyd/heap/contexts/")
		var isInShed = (relativePath == "dyd/shed") ||
			strings.HasPrefix(relativePath, "dyd/shed/") 
		var isInSprouts = (relativePath == "dyd/sprouts") ||
			strings.HasPrefix(relativePath, "dyd/sprouts/") 

		var shouldMatch = 
			isBase ||
			isDyd ||
			isTypeFile ||
			(isInRoots && req.IncludeRoots) ||
			(isInHeap && req.IncludeHeap) ||
			(isInContexts && req.IncludeContexts) ||
			(isInShed && req.IncludeShed) ||
			(isInSprouts && req.IncludeSprouts)
		shouldMatch = shouldMatch && !node.Info.IsDir()

		return nil, shouldMatch
	}

	err = fs2.DFSWalk3(
		ctx,
		fs2.Walk5Request{
			BasePath: gardenPath,
			Path: gardenPath,
			VPath: gardenPath,
			ShouldCrawl: packFilesShouldCrawl,
			ShouldMatch: packFilesShouldMatch,
			OnMatch:     packEntry,
		},
	)
	if err != nil {
		return err, ""
	}

	return err, targetPath
}
