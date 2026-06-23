package core

import (
	dydfs "dryad/filesystem"
	"dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/task"
	"strings"
)

type rootsWalkRequest struct {
	Roots       *SafeRootsReference
	ShouldMatch func(*task.ExecutionContext, *SafeRootReference) (error, bool)
	OnMatch     func(ctx *task.ExecutionContext, match *SafeRootReference) (error, any)
}

type rootsWalkIsRootKey struct {
	Group string
	Path  string
}

func rootsWalkIsRootCacheKey(path string) rootsWalkIsRootKey {
	return rootsWalkIsRootKey{
		Group: "RootsWalk.IsRoot",
		Path:  path,
	}
}

func rootsWalk(ctx *task.ExecutionContext, req rootsWalkRequest) (error, any) {
	var rootsPath = req.Roots.BasePath
	var err error

	var isRootCtx = &task.ExecutionContext{
		ConcurrencyChannel: task.DEFAULT_CONTEXT.ConcurrencyChannel,
	}
	if ctx != nil {
		isRootCtx.ConcurrencyChannel = ctx.ConcurrencyChannel
	}

	var isRoot = task.Memoize(
		func(ctx *task.ExecutionContext, path string) (error, bool) {
			typePath := filepath.Join(path, "dyd", "type")
			typeBytes, err := os.ReadFile(typePath)

			isRoot := err == nil && string(typeBytes) == "root"

			return nil, isRoot
		},
		func(ctx *task.ExecutionContext, path string) (error, any) {
			return nil, rootsWalkIsRootCacheKey(path)
		},
	)
	isRoot = task.WithContext(
		isRoot,
		func(ctx *task.ExecutionContext, path string) (error, *task.ExecutionContext) {
			return nil, isRootCtx
		},
	)

	var shouldWalk = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		err, root := isRoot(ctx, node.Path)
		return err, !root
	}

	var shouldMatch = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, bool) {
		// Only directories and symlinks can be roots; regular files cannot contain dyd/type.
		if !node.Info.IsDir() && node.Info.Mode()&os.ModeSymlink != os.ModeSymlink {
			return nil, false
		}

		err, root := isRoot(ctx, node.Path)
		return err, root
	}

	var onMatch = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
		var safeRootRef SafeRootReference
		var err error
		relPath, relErr := filepath.Rel(req.Roots.BasePath, node.Path)
		isInRootsPath := relErr == nil && relPath != ".." && !strings.HasPrefix(relPath, "../")

		if node.Info.Mode()&os.ModeSymlink == os.ModeSymlink || !isInRootsPath {
			unsafeRootRef := UnsafeRootReference{
				BasePath: node.Path,
				Roots:    req.Roots,
			}

			err, safeRootRef = unsafeRootRef.Resolve(ctx)
			if err != nil {
				return err, nil
			}
		} else {
			safeRootRef = SafeRootReference{
				BasePath: node.Path,
				Roots:    req.Roots,
			}
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

	var onPostMatch = func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, any) {
		ctx.ExecutionCache.Delete(rootsWalkIsRootCacheKey(node.Path))
		return nil, nil
	}
	onPostMatch = task.WithContext(
		onPostMatch,
		func(ctx *task.ExecutionContext, node dydfs.Walk6Node) (error, *task.ExecutionContext) {
			return nil, isRootCtx
		},
	)

	err, _ = dydfs.Walk6(
		ctx,
		dydfs.Walk6Request{
			BasePath:    rootsPath,
			Path:        rootsPath,
			VPath:       rootsPath,
			ShouldWalk:  shouldWalk,
			OnPreMatch:  onMatch,
			OnPostMatch: onPostMatch,
		},
	)

	return err, nil
}

type RootsWalkRequest struct {
	ShouldMatch func(*task.ExecutionContext, *SafeRootReference) (error, bool)
	OnMatch     func(*task.ExecutionContext, *SafeRootReference) (error, any)
}

func (roots *SafeRootsReference) Walk(ctx *task.ExecutionContext, req RootsWalkRequest) error {
	if req.ShouldMatch == nil {
		req.ShouldMatch = func(*task.ExecutionContext, *SafeRootReference) (error, bool) {
			return nil, true
		}
	}

	err, _ := rootsWalk(
		ctx,
		rootsWalkRequest{
			Roots:       roots,
			ShouldMatch: req.ShouldMatch,
			OnMatch:     req.OnMatch,
		},
	)
	return err
}
