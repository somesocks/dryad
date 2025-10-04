package core

import (
	dydfs "dryad/filesystem"
	"dryad/task"
	"os"
	"path/filepath"
)

type rootsWalkRequest struct {
	Roots *SafeRootsReference
	ShouldMatch func(*task.ExecutionContext, *SafeRootReference) (error, bool)
	OnMatch func (ctx *task.ExecutionContext, match *SafeRootReference) (error, any)
}

func rootsWalk(ctx *task.ExecutionContext, req rootsWalkRequest) (error, any) {
	var rootsPath = req.Roots.BasePath
	var err error

	var isRoot = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		typePath := filepath.Join(node.Path, "dyd", "type")
		typeBytes, err := os.ReadFile(typePath)
	
		isRoot := err == nil && string(typeBytes) == "root"
	
		return  nil, isRoot
	}
	
	var shouldWalk = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		err, isRoot := isRoot(ctx, node)
		return err, !isRoot
	}
	
	var shouldMatch = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		err, isRoot := isRoot(ctx, node)
		return err, isRoot
	}

	var onMatch = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
		var unsafeRootRef = UnsafeRootReference{
			BasePath: node.Path,
			Roots: req.Roots,
		}
		var safeRootRef SafeRootReference
		var err error

		err, safeRootRef = unsafeRootRef.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, shouldMatchRoot := req.ShouldMatch(ctx, &safeRootRef)
		if err != nil {
			return err, nil
		} else if !shouldMatchRoot {
			return nil, nil
		}

		err, _ = req.OnMatch(ctx, &safeRootRef)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	onMatch = dydfs.ConditionalWalkAction(onMatch, shouldMatch)

	err, _ = dydfs.Walk6(
		ctx,
		dydfs.Walk6Request{
			BasePath: rootsPath,
			Path: rootsPath,
			VPath: rootsPath,
			ShouldWalk: shouldWalk,
			OnPreMatch: onMatch,
		},
	)

	return err, nil
}

type RootsWalkRequest struct {
	ShouldMatch func(*task.ExecutionContext, *SafeRootReference) (error, bool)
	OnMatch func (*task.ExecutionContext, *SafeRootReference) (error, any)
}

func (roots *SafeRootsReference) Walk(ctx *task.ExecutionContext, req RootsWalkRequest) (error) {
	if req.ShouldMatch == nil {
		req.ShouldMatch = func(*task.ExecutionContext, *SafeRootReference) (error, bool) {
			return nil, true
		}
	}

	err, _ := rootsWalk(
		ctx,
		rootsWalkRequest{
			Roots: roots,
			ShouldMatch: req.ShouldMatch,
			OnMatch: req.OnMatch,
		},
	)
	return err
}