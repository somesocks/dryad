package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"
	"os"
	"path/filepath"
)

type RootsWalkRequest struct {
	Garden *SafeGardenReference
	OnRoot func (ctx *task.ExecutionContext, match RootsWalkMatch) (error, any)
}

type RootsWalkMatch struct {
	RootPath string
	GardenPath string
}

func RootsWalk(ctx *task.ExecutionContext, req RootsWalkRequest) (error, any) {
	var rootsPath, err = RootsPath(req.Garden)
	if err != nil {
		return err, nil
	}

	var isRoot = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		typePath := filepath.Join(node.Path, "dyd", "type")
		typeBytes, err := os.ReadFile(typePath)
	
		isRoot := err == nil && string(typeBytes) == "root"
	
		return  nil, isRoot
	}
	
	var shouldCrawl = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		err, isRoot := isRoot(ctx, node)
		return err, !isRoot
	}
	
	var shouldMatch = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
		err, isRoot := isRoot(ctx, node)
		return err, isRoot
	}

	var onMatch = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
		err, _ := req.OnRoot(ctx, RootsWalkMatch{
			GardenPath: req.Garden.BasePath,
			RootPath: node.Path,
		})
		return err, nil
	}

	err, _ = fs2.BFSWalk3(
		ctx,
		fs2.Walk5Request{
			Path: rootsPath,
			VPath: rootsPath,
			BasePath: rootsPath,
			ShouldCrawl: shouldCrawl,
			ShouldMatch: shouldMatch,
			OnMatch: onMatch,
		},
	)

	return err, nil
}
	
	
// 	path string, walkFn func(path string, info fs.FileInfo) error) error {
// 	var rootsPath, err = RootsPath(path)
// 	if err != nil {
// 		return err
// 	}

// 	err = fs2.Walk(fs2.WalkRequest{
// 		BasePath:     rootsPath,
// 		CrawlExclude: _isRoot,
// 		MatchInclude: _isRoot,
// 		OnMatch:      walkFn,
// 	})
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }
