package core

import "io/fs"

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

func RootsGraph(gardenPath string) (TRootsGraph, error) {
	graph := make(TRootsGraph)

	err := RootsWalk(gardenPath, func(rootPath string, info fs.FileInfo) error {
		err := RootRequirementsWalk(rootPath, func(requirementPath string, info fs.FileInfo) error {
			graph.AddEdge(rootPath, requirementPath)
			return nil
		})
		return err
	})

	if err != nil {
		return graph, err
	}

	return graph, nil
}
