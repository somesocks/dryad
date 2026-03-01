package core

import (
	"dryad/task"
	"fmt"
	"sort"
	"strings"
)

type rootBuildCommandsSelector struct {
	Name       string
	Path       string
	Descriptor VariantDescriptor
}

func rootBuild_readCommandsSelectors(rootPath string) (error, []rootBuildCommandsSelector) {
	err, _, commandsSelectors := rootBuild_readAssetsAndCommandsSelectors(rootPath)
	return err, commandsSelectors
}

func rootBuild_commandsSelectorMatchesVariant(
	dimensions []VariantDimension,
	selector VariantDescriptor,
	concreteVariant VariantDescriptor,
) (error, bool) {
	dimensionsByName := map[string]VariantDimension{}
	dimensionNames := make([]string, 0, len(dimensions))
	for _, dimension := range dimensions {
		dimensionsByName[dimension.Name] = dimension
		dimensionNames = append(dimensionNames, dimension.Name)
	}
	sort.Strings(dimensionNames)

	for selectorDimension := range selector {
		_, exists := dimensionsByName[selectorDimension]
		if !exists {
			return fmt.Errorf("over-specified commands variant dimension: %s", selectorDimension), false
		}
	}

	for _, dimensionName := range dimensionNames {
		requestedOption, hasRequestedOption := selector[dimensionName]
		if !hasRequestedOption {
			requestedOption = VariantOptionAny
		}

		err, choices := rootVariantFilterResolveChoicesForDimension(
			dimensionsByName[dimensionName],
			requestedOption,
			"commands",
		)
		if err != nil {
			return err, false
		}

		concreteOption, hasConcreteOption := concreteVariant[dimensionName]
		matchesDimension := false
		for _, choice := range choices {
			if choice.Omit {
				if !hasConcreteOption {
					matchesDimension = true
					break
				}
				continue
			}

			if hasConcreteOption && concreteOption == choice.Option {
				matchesDimension = true
				break
			}
		}

		if !matchesDimension {
			return nil, false
		}
	}

	return nil, true
}

func rootBuild_selectCommandsPathFromSelectors(
	dimensions []VariantDimension,
	concreteVariant VariantDescriptor,
	variantDescriptor string,
	selectors []rootBuildCommandsSelector,
) (error, string) {
	matchingSelectors := make([]rootBuildCommandsSelector, 0)
	for _, selector := range selectors {
		err, matchesVariant := rootBuild_commandsSelectorMatchesVariant(
			dimensions,
			selector.Descriptor,
			concreteVariant,
		)
		if err != nil {
			return err, ""
		}
		if !matchesVariant {
			continue
		}

		matchingSelectors = append(matchingSelectors, selector)
	}

	if len(matchingSelectors) == 0 {
		return nil, ""
	}

	if len(matchingSelectors) > 1 {
		selectorNames := make([]string, 0, len(matchingSelectors))
		for _, selector := range matchingSelectors {
			selectorNames = append(selectorNames, selector.Name)
		}
		sort.Strings(selectorNames)
		return fmt.Errorf(
			"multiple matching dyd/commands selectors for variant %s: %s",
			rootBuildLogVariantLabel(variantDescriptor),
			strings.Join(selectorNames, ", "),
		), ""
	}

	return nil, matchingSelectors[0].Path
}

func rootBuild_selectCommandsPath(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string) {
	err, variantContext := RootVariantContextFromFilesystem(variantDescriptor)
	if err != nil {
		return err, ""
	}

	rootRef := SafeRootReference{BasePath: rootPath}
	err, dimensions := rootRef.VariantDimensions(ctx)
	if err != nil {
		return err, ""
	}

	err, selectors := rootBuild_readCommandsSelectors(rootPath)
	if err != nil {
		return err, ""
	}

	return rootBuild_selectCommandsPathFromSelectors(
		dimensions,
		variantContext.Descriptor,
		variantDescriptor,
		selectors,
	)
}
