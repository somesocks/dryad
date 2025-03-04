package core

import (
	"path/filepath"

	"dryad/task"
)

type TStringSet map[string]bool

func (ss TStringSet) ToArray(results []string) []string {
	for key := range ss {
		results = append(results, key)
	}

	return results
}

type TRootsGraph map[string][]string

func (g TRootsGraph) AddEdge(src string, dest string) {
	g[src] = append(g[src], dest)
}

func (g TRootsGraph) Transpose() TRootsGraph {
	transposed := make(TRootsGraph)

	for src, neighbors := range g {
		for _, dest := range neighbors {
			transposed.AddEdge(dest, src)
		}
	}

	return transposed
}

func (g TRootsGraph) Descendants(results TStringSet, roots []string) TStringSet {
	if results == nil {
		results = make(map[string]bool)
	}

	for _, root := range roots {
		children := g[root]
		for _, child := range children {
			results[child] = true
		}
		g.Descendants(results, children)
	}

	return results
}

type rootsGraphRequest struct {
	Roots *SafeRootsReference
	Relative bool
}

func rootsGraph(
	ctx *task.ExecutionContext,
	req rootsGraphRequest,
) (error, TRootsGraph) {
	var err error
	var relative bool = req.Relative

	graph := make(TRootsGraph)

	var onRootRequirement = func (ctx *task.ExecutionContext, requirement *SafeRootRequirementReference) (error, any) {
		var rootPath string = requirement.Requirements.Root.BasePath
		var targetPath string
		var target *SafeRootReference
		var err error

		err, target = requirement.Target(ctx)
		if err != nil {
			return err, nil
		}
		
		targetPath = target.BasePath

		if relative {
			var gardenPath string = requirement.Requirements.Root.Roots.Garden.BasePath
			rootPath, err = filepath.Rel(gardenPath, rootPath)
			if err != nil {
				return err, nil
			}
			targetPath, err = filepath.Rel(gardenPath, targetPath)
			if err != nil {
				return err, nil
			}
		}

		graph.AddEdge(rootPath, targetPath)

		return nil, nil
	}

	var onRoot = func (ctx *task.ExecutionContext, root *SafeRootReference) (error, any) {
		var requirements *SafeRootRequirementsReference
		var err error

		err, requirements = root.Requirements().Resolve(ctx)
		if err != nil {
			return err, nil
		} else if requirements == nil {
			// do nothing if there are no requirements
			return nil, nil
		}

		err = requirements.Walk(
			ctx,
			RootRequirementsWalkRequest{
				OnMatch: onRootRequirement,
			},
		)
		if err != nil {
			return err, nil
		}

		return nil, nil
	}

	err = req.Roots.Walk(
		task.SERIAL_CONTEXT,
		RootsWalkRequest{
			OnMatch: onRoot,
		},
	)

	if err != nil {
		return err, graph
	}

	return nil, graph
}

type RootsGraphRequest struct {
	Relative bool
}

func (roots *SafeRootsReference) Graph(ctx *task.ExecutionContext, req RootsGraphRequest) (error, TRootsGraph) {
	err, graph := rootsGraph(
		ctx,
		rootsGraphRequest{
			Roots: roots,
			Relative: req.Relative,
		},
	)
	return err, graph
}