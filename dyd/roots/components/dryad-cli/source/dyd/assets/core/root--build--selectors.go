package core

import (
	"dryad/task"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type rootBuildSelectorMatchContext struct {
	DimensionsByName map[string]VariantDimension
	DimensionNames   []string
	ConcreteVariant  VariantDescriptor
}

type rootBuildSelectorSelectionState struct {
	MatchCount int
	MatchPath  string
}

func (state *rootBuildSelectorSelectionState) RecordMatch(path string) {
	if state.MatchCount == 0 {
		state.MatchPath = path
	}
	state.MatchCount++
}

func rootBuild_parseVariantSelectorDescriptor(
	selectorName string,
	baseName string,
	selectorLabel string,
) (error, bool, VariantDescriptor) {
	selectorRaw, hasSelector := strings.CutPrefix(
		selectorName,
		baseName+RootRequirementSelectorSeparator,
	)

	if selectorName == baseName {
		return nil, true, VariantDescriptor{}
	}
	if !hasSelector {
		return nil, false, nil
	}

	if selectorRaw == "" {
		return fmt.Errorf("malformed %s selector: %s", selectorLabel, selectorName), false, nil
	}

	err, parsedDescriptor := variantDescriptorParseFilesystem(selectorRaw)
	if err != nil {
		return fmt.Errorf("malformed %s selector: %s", selectorLabel, selectorName), false, nil
	}

	err, normalizedSelector := variantDescriptorEncodeFilesystem(parsedDescriptor)
	if err != nil {
		return err, false, nil
	}
	if normalizedSelector != selectorRaw {
		return fmt.Errorf("%s selector must be canonical: %s", selectorLabel, selectorName), false, nil
	}

	return nil, true, parsedDescriptor
}

func rootBuild_newSelectorMatchContext(
	dimensions []VariantDimension,
	concreteVariant VariantDescriptor,
) rootBuildSelectorMatchContext {
	dimensionsByName := map[string]VariantDimension{}
	dimensionNames := make([]string, 0, len(dimensions))
	for _, dimension := range dimensions {
		dimensionsByName[dimension.Name] = dimension
		dimensionNames = append(dimensionNames, dimension.Name)
	}
	sort.Strings(dimensionNames)

	return rootBuildSelectorMatchContext{
		DimensionsByName: dimensionsByName,
		DimensionNames:   dimensionNames,
		ConcreteVariant:  concreteVariant,
	}
}

func rootBuild_selectorMatchesVariant(
	ctx rootBuildSelectorMatchContext,
	selector VariantDescriptor,
	selectorType string,
) (error, bool) {
	for selectorDimension := range selector {
		_, exists := ctx.DimensionsByName[selectorDimension]
		if !exists {
			return fmt.Errorf("over-specified %s variant dimension: %s", selectorType, selectorDimension), false
		}
	}

	for _, dimensionName := range ctx.DimensionNames {
		requestedOption, hasRequestedOption := selector[dimensionName]
		if !hasRequestedOption {
			requestedOption = VariantOptionAny
		}

		err, choices := rootVariantFilterResolveChoicesForDimension(
			ctx.DimensionsByName[dimensionName],
			requestedOption,
			selectorType,
		)
		if err != nil {
			return err, false
		}

		concreteOption, hasConcreteOption := ctx.ConcreteVariant[dimensionName]
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

func rootBuild_considerSelectorMatch(
	matchContext rootBuildSelectorMatchContext,
	selectorName string,
	selectorPath string,
	baseName string,
	selectorLabel string,
	selectorType string,
	state *rootBuildSelectorSelectionState,
) error {
	err, isSelector, descriptor := rootBuild_parseVariantSelectorDescriptor(
		selectorName,
		baseName,
		selectorLabel,
	)
	if err != nil {
		return err
	}
	if !isSelector {
		return nil
	}

	err, matchesVariant := rootBuild_selectorMatchesVariant(
		matchContext,
		descriptor,
		selectorType,
	)
	if err != nil {
		return err
	}
	if matchesVariant {
		state.RecordMatch(selectorPath)
	}

	return nil
}

func rootBuild_selectAssetsAndCommandsAndSecretsPaths(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string, string, string) {
	err, variantContext := RootVariantContextFromFilesystem(variantDescriptor)
	if err != nil {
		return err, "", "", ""
	}

	rootRef := SafeRootReference{BasePath: rootPath}
	err, dimensions := rootRef.VariantDimensions(ctx)
	if err != nil {
		return err, "", "", ""
	}

	matchContext := rootBuild_newSelectorMatchContext(
		dimensions,
		variantContext.Descriptor,
	)

	dydPath := filepath.Join(rootPath, "dyd")
	dydPathExists, err := fileExists(dydPath)
	if err != nil {
		return err, "", "", ""
	}
	if !dydPathExists {
		return nil, "", "", ""
	}

	dydDir, err := os.Open(dydPath)
	if err != nil {
		return err, "", "", ""
	}
	defer dydDir.Close()

	assetsState := rootBuildSelectorSelectionState{}
	commandsState := rootBuildSelectorSelectionState{}
	secretsState := rootBuildSelectorSelectionState{}

	for {
		entryNames, err := dydDir.Readdirnames(64)
		if err != nil && err != io.EOF {
			return err, "", "", ""
		}

		for _, selectorName := range entryNames {
			selectorPath := filepath.Join(dydPath, selectorName)

			err = rootBuild_considerSelectorMatch(
				matchContext,
				selectorName,
				selectorPath,
				"assets",
				"dyd/assets",
				"assets",
				&assetsState,
			)
			if err != nil {
				return err, "", "", ""
			}

			err = rootBuild_considerSelectorMatch(
				matchContext,
				selectorName,
				selectorPath,
				"commands",
				"dyd/commands",
				"commands",
				&commandsState,
			)
			if err != nil {
				return err, "", "", ""
			}

			err = rootBuild_considerSelectorMatch(
				matchContext,
				selectorName,
				selectorPath,
				"secrets",
				"dyd/secrets",
				"secrets",
				&secretsState,
			)
			if err != nil {
				return err, "", "", ""
			}
		}

		if err == io.EOF {
			break
		}
	}

	variantLabel := rootBuildLogVariantLabel(variantDescriptor)
	if assetsState.MatchCount > 1 {
		return fmt.Errorf("multiple matching dyd/assets selectors for variant %s", variantLabel), "", "", ""
	}
	if commandsState.MatchCount > 1 {
		return fmt.Errorf("multiple matching dyd/commands selectors for variant %s", variantLabel), "", "", ""
	}
	if secretsState.MatchCount > 1 {
		return fmt.Errorf("multiple matching dyd/secrets selectors for variant %s", variantLabel), "", "", ""
	}

	return nil, assetsState.MatchPath, commandsState.MatchPath, secretsState.MatchPath
}
