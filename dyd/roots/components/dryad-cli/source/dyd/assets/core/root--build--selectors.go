package core

import (
	"dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/task"
	"fmt"
	"io"
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

type rootVariantSelectedPathValues struct {
	AssetsPath       string
	CommandsPath     string
	TraitsPath       string
	SecretsPath      string
	DocsPath         string
	RequirementsPath string
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

		err, choices := rootVariantFilterResolveChoicesForDimension(
			ctx.DimensionsByName[dimensionName],
			requestedOption,
			hasRequestedOption,
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

func rootVariant_resolveConcreteDescriptor(
	ctx *task.ExecutionContext,
	variant *UnsafeRootVariantReference,
) (error, VariantDescriptor, []VariantDimension) {
	err, normalizedDescriptor := normalizeVariantDescriptor(variant.Descriptor)
	if err != nil {
		return err, nil, nil
	}

	err, dimensions := variant.Root.VariantDimensions(ctx)
	if err != nil {
		return err, nil, nil
	}

	dimensionsByName := map[string]VariantDimension{}
	for _, dimension := range dimensions {
		dimensionsByName[dimension.Name] = dimension
	}

	concreteDescriptor := VariantDescriptor{}

	for descriptorDimension := range normalizedDescriptor {
		_, exists := dimensionsByName[descriptorDimension]
		if !exists {
			return fmt.Errorf("over-specified root variant descriptor dimension: %s", descriptorDimension), nil, nil
		}
	}

	for _, dimension := range dimensions {
		optionByName := map[string]VariantDimensionOption{}
		for _, option := range dimension.Options {
			optionByName[option.Name] = option
		}

		selectedOption, specified := normalizedDescriptor[dimension.Name]
		if !specified {
			if noneOption, hasNone := optionByName[VariantOptionNone]; hasNone && noneOption.Enabled {
				continue
			}
			return fmt.Errorf("under-specified root variant descriptor dimension: %s", dimension.Name), nil, nil
		}

		switch selectedOption {
		case VariantOptionInherit, VariantOptionAny, VariantOptionHost:
			return fmt.Errorf("invalid concrete root variant option for %s: %s", dimension.Name, selectedOption), nil, nil
		}

		option, exists := optionByName[selectedOption]
		if !exists {
			return fmt.Errorf("wrongly-specified root variant option: %s=%s", dimension.Name, selectedOption), nil, nil
		}
		if !option.Enabled {
			return fmt.Errorf("disabled root variant option: %s=%s", dimension.Name, selectedOption), nil, nil
		}
		if selectedOption == VariantOptionNone {
			continue
		}

		concreteDescriptor[dimension.Name] = selectedOption
	}

	return nil, concreteDescriptor, dimensions
}

func rootVariant_resolveSelectedPathValues(
	ctx *task.ExecutionContext,
	root *SafeRootReference,
	concreteDescriptor VariantDescriptor,
	dimensions []VariantDimension,
) (error, rootVariantSelectedPathValues) {
	matchContext := rootBuild_newSelectorMatchContext(
		dimensions,
		concreteDescriptor,
	)

	dydPath := filepath.Join(root.BasePath, "dyd")
	dydDir, err := os.Open(dydPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, rootVariantSelectedPathValues{}
		}
		return err, rootVariantSelectedPathValues{}
	}
	defer dydDir.Close()

	assetsState := rootBuildSelectorSelectionState{}
	commandsState := rootBuildSelectorSelectionState{}
	traitsState := rootBuildSelectorSelectionState{}
	secretsState := rootBuildSelectorSelectionState{}
	docsState := rootBuildSelectorSelectionState{}
	requirementsState := rootBuildSelectorSelectionState{}

	for {
		entryNames, err := dydDir.Readdirnames(64)
		if err != nil && err != io.EOF {
			return err, rootVariantSelectedPathValues{}
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
				return err, rootVariantSelectedPathValues{}
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
				return err, rootVariantSelectedPathValues{}
			}

			err = rootBuild_considerSelectorMatch(
				matchContext,
				selectorName,
				selectorPath,
				"traits",
				"dyd/traits",
				"traits",
				&traitsState,
			)
			if err != nil {
				return err, rootVariantSelectedPathValues{}
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
				return err, rootVariantSelectedPathValues{}
			}

			err = rootBuild_considerSelectorMatch(
				matchContext,
				selectorName,
				selectorPath,
				"docs",
				"dyd/docs",
				"docs",
				&docsState,
			)
			if err != nil {
				return err, rootVariantSelectedPathValues{}
			}

			err = rootBuild_considerSelectorMatch(
				matchContext,
				selectorName,
				selectorPath,
				"requirements",
				"dyd/requirements",
				"requirements",
				&requirementsState,
			)
			if err != nil {
				return err, rootVariantSelectedPathValues{}
			}
		}

		if err == io.EOF {
			break
		}
	}

	err, variantLabel := variantDescriptorEncodeFilesystem(concreteDescriptor)
	if err != nil {
		return err, rootVariantSelectedPathValues{}
	}

	variantLabel = rootBuildLogVariantLabel(variantLabel)
	if assetsState.MatchCount > 1 {
		return fmt.Errorf("multiple matching dyd/assets selectors for variant %s", variantLabel), rootVariantSelectedPathValues{}
	}
	if commandsState.MatchCount > 1 {
		return fmt.Errorf("multiple matching dyd/commands selectors for variant %s", variantLabel), rootVariantSelectedPathValues{}
	}
	if traitsState.MatchCount > 1 {
		return fmt.Errorf("multiple matching dyd/traits selectors for variant %s", variantLabel), rootVariantSelectedPathValues{}
	}
	if secretsState.MatchCount > 1 {
		return fmt.Errorf("multiple matching dyd/secrets selectors for variant %s", variantLabel), rootVariantSelectedPathValues{}
	}
	if docsState.MatchCount > 1 {
		return fmt.Errorf("multiple matching dyd/docs selectors for variant %s", variantLabel), rootVariantSelectedPathValues{}
	}
	if requirementsState.MatchCount > 1 {
		return fmt.Errorf("multiple matching dyd/requirements selectors for variant %s", variantLabel), rootVariantSelectedPathValues{}
	}

	return nil, rootVariantSelectedPathValues{
		AssetsPath:       assetsState.MatchPath,
		CommandsPath:     commandsState.MatchPath,
		TraitsPath:       traitsState.MatchPath,
		SecretsPath:      secretsState.MatchPath,
		DocsPath:         docsState.MatchPath,
		RequirementsPath: requirementsState.MatchPath,
	}
}

func (variant *UnsafeRootVariantReference) Resolve(
	ctx *task.ExecutionContext,
) (error, *SafeRootVariantReference) {
	err, concreteDescriptor, dimensions := rootVariant_resolveConcreteDescriptor(ctx, variant)
	if err != nil {
		return err, nil
	}

	safeVariant := &SafeRootVariantReference{
		Root:       variant.Root,
		Descriptor: concreteDescriptor,
		Dimensions: dimensions,
	}

	err, selectedPathValues := rootVariant_resolveSelectedPathValues(
		ctx,
		variant.Root,
		concreteDescriptor,
		dimensions,
	)
	if err != nil {
		return err, nil
	}

	safeVariant.Assets = rootVariantSelectedAssets(selectedPathValues.AssetsPath, safeVariant)
	safeVariant.Commands = rootVariantSelectedCommands(selectedPathValues.CommandsPath, safeVariant)
	safeVariant.Traits = rootVariantSelectedTraits(selectedPathValues.TraitsPath, safeVariant)
	safeVariant.Secrets = rootVariantSelectedSecrets(selectedPathValues.SecretsPath, safeVariant)
	safeVariant.Docs = rootVariantSelectedDocs(selectedPathValues.DocsPath, safeVariant)
	safeVariant.Requirements = rootVariantSelectedRequirements(selectedPathValues.RequirementsPath, safeVariant)

	return nil, safeVariant
}
