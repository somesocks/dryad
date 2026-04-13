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
	if idx := strings.LastIndex(rootPath, RootRequirementSelectorSeparator); idx > -1 && idx < len(rootPath)-len(RootRequirementSelectorSeparator) {
		selectorRaw := rootPath[idx+len(RootRequirementSelectorSeparator):]
		if err, _ := variantDescriptorNormalizeFilesystem(selectorRaw); err == nil {
			rootPath = rootPath[:idx]
		}
	}
	return rootPath
}

func rootGraphNode(path string, descriptor VariantDescriptor) (error, string) {
	err, selectorRaw := variantDescriptorEncodeFilesystem(descriptor)
	if err != nil {
		return err, ""
	}
	if selectorRaw == "" {
		return nil, path
	}
	return nil, path + RootRequirementSelectorSeparator + selectorRaw
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
		transposed.EnsureNode(src)

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

func (g TRootsGraph) descendantNodesFromNode(results TStringSet, visited map[string]bool, node string) {
	if visited[node] {
		return
	}
	visited[node] = true

	for _, target := range g[node] {
		results[target] = true
		g.descendantNodesFromNode(results, visited, target)
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

func (g TRootsGraph) DescendantNodes(results TStringSet, nodes []string) TStringSet {
	if results == nil {
		results = make(map[string]bool)
	}
	visited := map[string]bool{}

	for _, node := range nodes {
		g.descendantNodesFromNode(results, visited, node)
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
	var relative bool = req.Relative

	graph := make(TRootsGraph)
	var graphMux sync.Mutex

	var onVariant = func(ctx *task.ExecutionContext, sourceVariant *SafeRootVariantReference) (error, any) {
		var rootPath string
		var root *SafeRootReference = sourceVariant.Root
		var gardenPath string = root.Roots.Garden.BasePath

		rootPath = root.BasePath
		if relative {
			var err error
			rootPath, err = filepath.Rel(gardenPath, rootPath)
			if err != nil {
				return err, nil
			}
		}

		err, sourceNode := rootGraphNode(rootPath, sourceVariant.Descriptor)
		if err != nil {
			return err, nil
		}

		graphMux.Lock()
		graph.EnsureNode(sourceNode)
		graphMux.Unlock()

		requirements := sourceVariant.Requirements
		if requirements == nil {
			return nil, nil
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

				err, shouldInclude := rootRequirementConditionMatches(sourceVariant.Descriptor, condition)
				if err != nil {
					return err, nil
				}
				if !shouldInclude {
					return nil, nil
				}

				err, targets := requirement.ResolveTargets(ctx, RootRequirementResolveTargetsRequest{
					ParentVariant: sourceVariant.Descriptor,
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

					err, targetNode := rootGraphNode(targetRootPath, target.VariantDescriptor)
					if err != nil {
						return err, nil
					}

					graphMux.Lock()
					graph.AddEdge(sourceNode, edgeName, targetNode)
					graphMux.Unlock()
				}

				return nil, nil
			},
		})
		return err, nil
	}

	err := req.Roots.WalkVariants(
		ctx,
		RootsWalkVariantsRequest{
			OnMatch: onVariant,
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
