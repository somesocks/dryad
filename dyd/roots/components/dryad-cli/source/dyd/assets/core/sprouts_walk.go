package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"
	"os"
	"path/filepath"
)

var sproutsWalk_ShouldCrawl = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
	isSymlink := node.Info.Mode()&os.ModeSymlink == os.ModeSymlink
	isDir := node.Info.IsDir()
	return nil, isDir && !isSymlink
}

var sproutsWalk_ShouldMatch = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, bool) {
	var dydPath = filepath.Join(node.Path, "dyd", "fingerprint")
	var _, dydInfoErr = os.Stat(dydPath)
	var isSprout = dydInfoErr == nil

	return nil, isSprout
}

type SproutsWalkRequest struct {
	GardenPath string
	OnSprout func(*task.ExecutionContext, string) (error, any)
}

func SproutsWalk(ctx *task.ExecutionContext, req SproutsWalkRequest) (error, any) {
	var sproutsPath, err = SproutsPath(req.GardenPath)
	if err != nil {
		return err, nil
	}

	var onMatch = func(ctx *task.ExecutionContext, node fs2.Walk5Node) (error, any) {
		err, _ := req.OnSprout(ctx, node.Path)
		return err, nil
	}

	err, _ = fs2.BFSWalk3(
		ctx,
		fs2.Walk5Request{
			BasePath:     sproutsPath,
			Path:     sproutsPath,
			VPath:     sproutsPath,
			ShouldCrawl: sproutsWalk_ShouldCrawl,
			ShouldMatch: sproutsWalk_ShouldMatch,
			OnMatch:      onMatch,
		},
	)
	if err != nil {
		return err, nil
	}

	return nil, nil
}
