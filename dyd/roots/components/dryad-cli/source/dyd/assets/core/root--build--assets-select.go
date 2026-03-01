package core

import (
	"dryad/task"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type rootBuildAssetsSelector struct {
	Name       string
	Path       string
	Descriptor VariantDescriptor
}

func rootBuild_readAssetsSelectors(rootPath string) (error, []rootBuildAssetsSelector) {
	dydPath := filepath.Join(rootPath, "dyd")
	dydPathExists, err := fileExists(dydPath)
	if err != nil {
		return err, nil
	}
	if !dydPathExists {
		return nil, []rootBuildAssetsSelector{}
	}

	dydEntries, err := os.ReadDir(dydPath)
	if err != nil {
		return err, nil
	}
	sort.Slice(dydEntries, func(i int, j int) bool {
		return dydEntries[i].Name() < dydEntries[j].Name()
	})

	selectors := make([]rootBuildAssetsSelector, 0)
	for _, dydEntry := range dydEntries {
		selectorName := dydEntry.Name()
		selectorRaw, hasSelector := strings.CutPrefix(
			selectorName,
			"assets"+RootRequirementSelectorSeparator,
		)

		descriptor := VariantDescriptor{}
		if selectorName == "assets" {
			descriptor = VariantDescriptor{}
		} else if hasSelector {
			if selectorRaw == "" {
				return fmt.Errorf("malformed dyd/assets selector: %s", selectorName), nil
			}

			err, parsedDescriptor := variantDescriptorParseFilesystem(selectorRaw)
			if err != nil {
				return fmt.Errorf("malformed dyd/assets selector: %s", selectorName), nil
			}

			err, normalizedSelector := variantDescriptorEncodeFilesystem(parsedDescriptor)
			if err != nil {
				return err, nil
			}
			if normalizedSelector != selectorRaw {
				return fmt.Errorf("dyd/assets selector must be canonical: %s", selectorName), nil
			}

			descriptor = parsedDescriptor
		} else {
			continue
		}

		selectors = append(selectors, rootBuildAssetsSelector{
			Name:       selectorName,
			Path:       filepath.Join(dydPath, selectorName),
			Descriptor: descriptor,
		})
	}

	return nil, selectors
}

func rootBuild_assetsSelectorMatchesVariant(
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
			return fmt.Errorf("over-specified assets variant dimension: %s", selectorDimension), false
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
			"assets",
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

func rootBuild_selectAssetsPath(
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

	err, selectors := rootBuild_readAssetsSelectors(rootPath)
	if err != nil {
		return err, ""
	}

	matchingSelectors := make([]rootBuildAssetsSelector, 0)
	for _, selector := range selectors {
		err, matchesVariant := rootBuild_assetsSelectorMatchesVariant(
			dimensions,
			selector.Descriptor,
			variantContext.Descriptor,
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
			"multiple matching dyd/assets selectors for variant %s: %s",
			rootBuildLogVariantLabel(variantDescriptor),
			strings.Join(selectorNames, ", "),
		), ""
	}

	return nil, matchingSelectors[0].Path
}
