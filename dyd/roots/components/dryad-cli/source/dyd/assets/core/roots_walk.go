package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"
	"os"
	"path/filepath"
)

type rootsWalkRequest struct {
	Roots *SafeRootsReference
	OnMatch func (ctx *task.ExecutionContext, match *SafeRootReference) (error, any)
}

func rootsWalk(ctx *task.ExecutionContext, req rootsWalkRequest) (error, any) {
	var rootsPath = req.Roots.BasePath
	var err error

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
		var unsafeRequirementRef = UnsafeRootReference{
			BasePath: node.Path,
			Roots: req.Roots,
		}
		var safeRequirementRef SafeRootReference
		var err error

		err, safeRequirementRef = unsafeRequirementRef.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, _ = req.OnMatch(ctx, &safeRequirementRef)
		if err != nil {
			return err, nil
		}

		return nil, nil
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

type RootsWalkRequest struct {
	OnMatch func (ctx *task.ExecutionContext, match *SafeRootReference) (error, any)
}

func (roots *SafeRootsReference) Walk(ctx *task.ExecutionContext, req RootsWalkRequest) (error) {
	err, _ := rootsWalk(
		ctx,
		rootsWalkRequest{
			Roots: roots,
			OnMatch: req.OnMatch,
		},
	)
	return err
}