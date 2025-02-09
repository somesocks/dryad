package core

import (
	"io/fs"
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

type RootsGraphRequest struct {
	Garden *SafeGardenReference
	Relative bool
}

func RootsGraph(req RootsGraphRequest) (TRootsGraph, error) {
	var err error
	var relative bool = req.Relative
	var gardenPath string = req.Garden.BasePath

	graph := make(TRootsGraph)

	err, _ = RootsWalk(
		task.SERIAL_CONTEXT,
		RootsWalkRequest{
			Garden: req.Garden,
			OnRoot : func (ctx *task.ExecutionContext, match RootsWalkMatch) (error, any) {
				rootPath, err := filepath.EvalSymlinks(match.RootPath)
				if err != nil {
					return err, nil
				}

				err = RootRequirementsWalk(rootPath, func(requirementPath string, info fs.FileInfo) error {
					requirementPath, err := filepath.EvalSymlinks(requirementPath)
					if err != nil {
						return err
					}
		
					if relative {
						relRootPath, err := filepath.Rel(gardenPath, rootPath)
						if err != nil {
							return err
						}
		
						relRequirementPath, err := filepath.Rel(gardenPath, requirementPath)
						if err != nil {
							return err
						}
		
						graph.AddEdge(relRootPath, relRequirementPath)
					} else {
						graph.AddEdge(rootPath, requirementPath)
					}
		
					return nil
				})

				if err != nil {
					return err, nil
				}
		
				return nil, nil
						

			},
		},
	)

	if err != nil {
		return graph, err
	}

	return graph, nil
}
