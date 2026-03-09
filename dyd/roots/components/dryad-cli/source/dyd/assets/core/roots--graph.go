package core

import (
	"dryad/internal/filepath"
	"sort"
	"strings"
	"sync"

	"dryad/task"
)

type TStringSet map[string]bool

func (ss TStringSet) ToArray(results []string) []string {
	for key := range ss {
		results = append(results, key)
	}

	return results
}

type TRootsGraph map[string]map[string]string

func (g TRootsGraph) EnsureNode(node string) {
	if g[node] == nil {
		g[node] = make(map[string]string)
	}
}

func (g TRootsGraph) AddEdge(src string, requirementName string, target string) {
	g.EnsureNode(src)
	g[src][requirementName] = target
}

func rootPathFromGraphNode(node string) string {
	rootPath := node
	if idx := strings.Index(rootPath, "?"); idx > -1 {
		rootPath = rootPath[:idx]
	}
	return rootPath
}

func (g TRootsGraph) nodesForRoot(root string) []string {
	nodes := []string{}

	if _, exists := g[root]; exists {
		nodes = append(nodes, root)
	}

	for node := range g {
		if node == root {
			continue
		}
		if rootPathFromGraphNode(node) == root {
			nodes = append(nodes, node)
		}
	}

	sort.Strings(nodes)
	return nodes
}

func (g TRootsGraph) Transpose() TRootsGraph {
	transposed := make(TRootsGraph)

	for src, requirements := range g {
		for requirementName, target := range requirements {
			transposedName := requirementName
			if transposed[target] != nil {
				if _, exists := transposed[target][transposedName]; exists {
					transposedName = src + "+" + requirementName
				}
			}
			transposed.AddEdge(target, transposedName, src)
		}
	}

	return transposed
}

func (g TRootsGraph) descendantsFromNode(results TStringSet, visited map[string]bool, node string) {
	if visited[node] {
		return
	}
	visited[node] = true

	for _, target := range g[node] {
		targetRoot := rootPathFromGraphNode(target)
		results[targetRoot] = true
		for _, nextNode := range g.nodesForRoot(targetRoot) {
			g.descendantsFromNode(results, visited, nextNode)
		}
	}
}

func (g TRootsGraph) Descendants(results TStringSet, roots []string) TStringSet {
	if results == nil {
		results = make(map[string]bool)
	}
	visited := map[string]bool{}

	for _, root := range roots {
		startNodes := g.nodesForRoot(root)
		if len(startNodes) == 0 {
			continue
		}
		for _, startNode := range startNodes {
			g.descendantsFromNode(results, visited, startNode)
		}
	}

	return results
}

type rootsGraphRequest struct {
	Roots    *SafeRootsReference
	Relative bool
}

func rootsGraph(
	ctx *task.ExecutionContext,
	req rootsGraphRequest,
) (error, TRootsGraph) {
	var err error
	var relative bool = req.Relative

	graph := make(TRootsGraph)
	var graphMux sync.Mutex

	var onRoot = func(ctx *task.ExecutionContext, root *SafeRootReference) (error, any) {
		var requirements *SafeRootRequirementsReference
		var err error
		var rootPath string
		var gardenPath string = root.Roots.Garden.BasePath

		rootPath = root.BasePath
		if relative {
			rootPath, err = filepath.Rel(gardenPath, rootPath)
			if err != nil {
				return err, nil
			}
		}

		err, sourceVariants := root.ResolveBuildVariants(ctx, RootResolveBuildVariantsRequest{
			Selector:                VariantDescriptor{},
			IgnoreUnknownDimensions: true,
		})
		if err != nil {
			return err, nil
		}

		err, requirements = root.Requirements().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		for _, sourceVariant := range sourceVariants {
			sourceVariant := sourceVariant

			err, sourceSelectorRaw := variantDescriptorEncodeURL(sourceVariant)
			if err != nil {
				return err, nil
			}
			sourceNode := rootPath + sourceSelectorRaw

			graphMux.Lock()
			graph.EnsureNode(sourceNode)
			graphMux.Unlock()

			if requirements == nil {
				continue
			}

			err = requirements.Walk(task.SERIAL_CONTEXT, RootRequirementsWalkRequest{
				OnMatch: func(ctx *task.ExecutionContext, requirement *SafeRootRequirementReference) (error, any) {
					requirementNameRaw := filepath.Base(requirement.BasePath)

					err, requirementName := RootRequirementNormalizeName(requirementNameRaw)
					if err != nil {
						return err, nil
					}

					err, _, condition := rootRequirementParseName(requirementName)
					if err != nil {
						return err, nil
					}

					err, shouldInclude := rootRequirementConditionMatches(sourceVariant, condition)
					if err != nil {
						return err, nil
					}
					if !shouldInclude {
						return nil, nil
					}

					err, targets := requirement.ResolveTargets(ctx, RootRequirementResolveTargetsRequest{
						ParentVariant: sourceVariant,
					})
					if err != nil {
						return err, nil
					}

					for _, target := range targets {
						err, edgeName := rootBuild_stage1DependencyName(requirementName, target, len(targets))
						if err != nil {
							return err, nil
						}

						targetRootPath := target.Root.BasePath
						if relative {
							targetRootPath, err = filepath.Rel(gardenPath, targetRootPath)
							if err != nil {
								return err, nil
							}
						}

						err, targetSelectorRaw := variantDescriptorEncodeURL(target.VariantDescriptor)
						if err != nil {
							return err, nil
						}
						targetNode := targetRootPath + targetSelectorRaw

						graphMux.Lock()
						graph.AddEdge(sourceNode, edgeName, targetNode)
						graphMux.Unlock()
					}

					return nil, nil
				},
			})
			if err != nil {
				return err, nil
			}
		}

		return nil, nil
	}

	err = req.Roots.Walk(
		ctx,
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
			Roots:    roots,
			Relative: req.Relative,
		},
	)
	return err, graph
}
