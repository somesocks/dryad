package core

import (
	"io/fs"
	"path/filepath"
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

func RootsGraph(gardenPath string, relative bool) (TRootsGraph, error) {
	gardenPath, err := GardenPath(gardenPath)
	if err != nil {
		return nil, err
	}

	graph := make(TRootsGraph)

	err = RootsWalk(gardenPath, func(rootPath string, info fs.FileInfo) error {
		rootPath, err := filepath.EvalSymlinks(rootPath)
		if err != nil {
			return err
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
			return err
		}

		return nil
	})

	if err != nil {
		return graph, err
	}

	return graph, nil
}
