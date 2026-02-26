package core

import (
	"dryad/task"
	"fmt"
	"runtime"
	"sort"
)

const (
	VariantOptionNone    = "none"
	VariantOptionInherit = "inherit"
	VariantOptionAny     = "any"
	VariantOptionHost    = "host"
)

type RootRequirementResolveVariantsRequest struct {
	ParentVariant VariantDescriptor
}

type RootRequirementResolveTargetsRequest struct {
	ParentVariant VariantDescriptor
}

type RootRequirementResolvedTarget struct {
	Root               *SafeRootReference
	VariantDescriptor  VariantDescriptor
	ForceVariantSuffix bool
}

type rootRequirementOptionChoice struct {
	Omit   bool
	Option string
}

func cloneVariantDescriptor(descriptor VariantDescriptor) VariantDescriptor {
	copy := VariantDescriptor{}
	for key, value := range descriptor {
		copy[key] = value
	}
	return copy
}

func normalizeVariantDescriptor(descriptor VariantDescriptor) (error, VariantDescriptor) {
	err, encoded := variantDescriptorEncodeFilesystem(descriptor)
	if err != nil {
		return err, nil
	}
	return variantDescriptorParseFilesystem(encoded)
}

func rootRequirementHostOption(dimensionName string) (error, string) {
	switch dimensionName {
	case "os":
		return nil, runtime.GOOS
	case "arch":
		return nil, runtime.GOARCH
	default:
		return fmt.Errorf("host option is only supported for variant dimensions os/arch: %s", dimensionName), ""
	}
}

func rootRequirementResolveChoicesForDimension(
	dimension VariantDimension,
	requestedOptionRaw string,
	parentVariant VariantDescriptor,
) (error, []rootRequirementOptionChoice) {
	exists := map[string]bool{}
	enabled := map[string]bool{}
	choices := make([]rootRequirementOptionChoice, 0)
	seenChoices := map[string]struct{}{}

	for _, option := range dimension.Options {
		exists[option.Name] = true
		enabled[option.Name] = option.Enabled
	}

	requireEnabledOption := func(optionName string) (error, rootRequirementOptionChoice) {
		if !exists[optionName] {
			return fmt.Errorf("wrongly-specified requirement variant option: %s=%s", dimension.Name, optionName), rootRequirementOptionChoice{}
		}
		if !enabled[optionName] {
			return fmt.Errorf("disabled requirement variant option: %s=%s", dimension.Name, optionName), rootRequirementOptionChoice{}
		}

		if optionName == VariantOptionNone {
			return nil, rootRequirementOptionChoice{Omit: true}
		}

		return nil, rootRequirementOptionChoice{Option: optionName}
	}

	appendUniqueChoice := func(choice rootRequirementOptionChoice) {
		key := choice.Option
		if choice.Omit {
			key = VariantOptionNone
		}
		if _, exists := seenChoices[key]; exists {
			return
		}
		seenChoices[key] = struct{}{}
		choices = append(choices, choice)
	}

	err, requestedOptions := variantDescriptorOptionValues(requestedOptionRaw)
	if err != nil {
		return err, nil
	}

	for _, requestedOption := range requestedOptions {
		switch requestedOption {
		case VariantOptionInherit:
			inheritedOption, hasInheritedOption := parentVariant[dimension.Name]
			if !hasInheritedOption {
				inheritedOption = VariantOptionNone
			}

			err, choice := requireEnabledOption(inheritedOption)
			if err != nil {
				return err, nil
			}
			appendUniqueChoice(choice)

		case VariantOptionAny:
			for _, option := range dimension.Options {
				if !option.Enabled {
					continue
				}

				if option.Name == VariantOptionNone {
					appendUniqueChoice(rootRequirementOptionChoice{Omit: true})
					continue
				}

				appendUniqueChoice(rootRequirementOptionChoice{Option: option.Name})
			}

			if len(choices) == 0 {
				return fmt.Errorf("no enabled variant options for any resolution: %s", dimension.Name), nil
			}

		case VariantOptionHost:
			err, hostOption := rootRequirementHostOption(dimension.Name)
			if err != nil {
				return err, nil
			}

			err, choice := requireEnabledOption(hostOption)
			if err != nil {
				return err, nil
			}
			appendUniqueChoice(choice)

		default:
			err, choice := requireEnabledOption(requestedOption)
			if err != nil {
				return err, nil
			}
			appendUniqueChoice(choice)
		}
	}

	return nil, choices
}

func (targetSpec *RootRequirementTargetSpec) ResolveVariants(
	ctx *task.ExecutionContext,
	req RootRequirementResolveVariantsRequest,
) (error, []VariantDescriptor) {
	err, parentVariant := normalizeVariantDescriptor(req.ParentVariant)
	if err != nil {
		return err, nil
	}

	err, requirementVariant := normalizeVariantDescriptor(targetSpec.VariantSelector)
	if err != nil {
		return err, nil
	}

	err, dimensions := targetSpec.Root.VariantDimensions(ctx)
	if err != nil {
		return err, nil
	}
	err, inclusions := targetSpec.Root.VariantInclusions(ctx)
	if err != nil {
		return err, nil
	}
	err, exclusions := targetSpec.Root.VariantExclusions(ctx)
	if err != nil {
		return err, nil
	}

	dimensionsByName := map[string]VariantDimension{}
	dimensionNames := make([]string, 0, len(dimensions))
	for _, dimension := range dimensions {
		dimensionsByName[dimension.Name] = dimension
		dimensionNames = append(dimensionNames, dimension.Name)
	}
	sort.Strings(dimensionNames)

	for requirementDimension := range requirementVariant {
		_, exists := dimensionsByName[requirementDimension]
		if !exists {
			return fmt.Errorf("over-specified requirement variant dimension: %s", requirementDimension), nil
		}
	}

	resolvedVariants := []VariantDescriptor{{}}
	for _, dimensionName := range dimensionNames {
		dimension := dimensionsByName[dimensionName]

		requestedOption, hasRequestedOption := requirementVariant[dimensionName]
		if !hasRequestedOption {
			return fmt.Errorf("under-specified requirement variant dimension: %s", dimensionName), nil
		}

		err, choices := rootRequirementResolveChoicesForDimension(
			dimension,
			requestedOption,
			parentVariant,
		)
		if err != nil {
			return err, nil
		}

		nextResolvedVariants := make([]VariantDescriptor, 0, len(resolvedVariants)*len(choices))
		for _, baseVariant := range resolvedVariants {
			for _, choice := range choices {
				nextVariant := cloneVariantDescriptor(baseVariant)
				if !choice.Omit {
					nextVariant[dimensionName] = choice.Option
				}
				nextResolvedVariants = append(nextResolvedVariants, nextVariant)
			}
		}

		resolvedVariants = nextResolvedVariants
	}

	byKey := map[string]VariantDescriptor{}
	keys := make([]string, 0, len(resolvedVariants))
	for _, resolvedVariant := range resolvedVariants {
		err, key := variantDescriptorEncodeFilesystem(resolvedVariant)
		if err != nil {
			return err, nil
		}
		if _, exists := byKey[key]; exists {
			continue
		}
		byKey[key] = resolvedVariant
		keys = append(keys, key)
	}
	sort.Strings(keys)

	results := make([]VariantDescriptor, 0, len(keys))
	for _, key := range keys {
		results = append(results, byKey[key])
	}

	err, results = applyVariantFilters(dimensions, inclusions, exclusions, results)
	if err != nil {
		return err, nil
	}
	if len(results) == 0 {
		return fmt.Errorf("resolved requirement variants are filtered by variants/_include and variants/_exclude"), nil
	}

	return nil, results
}

func (rootRequirement *SafeRootRequirementReference) ResolveTargets(
	ctx *task.ExecutionContext,
	req RootRequirementResolveTargetsRequest,
) (error, []RootRequirementResolvedTarget) {
	err, targetSpec := rootRequirement.TargetSpec(ctx)
	if err != nil {
		return err, nil
	}

	err, variants := targetSpec.ResolveVariants(ctx, RootRequirementResolveVariantsRequest{
		ParentVariant: req.ParentVariant,
	})
	if err != nil {
		return err, nil
	}

	forceVariantSuffix := false
	for _, optionRaw := range targetSpec.VariantSelector {
		err, options := variantDescriptorOptionValues(optionRaw)
		if err != nil {
			return err, nil
		}
		if len(options) > 1 {
			forceVariantSuffix = true
			break
		}
		for _, option := range options {
			if option == VariantOptionAny {
				forceVariantSuffix = true
				break
			}
		}
		if forceVariantSuffix {
			break
		}
	}

	results := make([]RootRequirementResolvedTarget, 0, len(variants))
	for _, variant := range variants {
		results = append(results, RootRequirementResolvedTarget{
			Root:               targetSpec.Root,
			VariantDescriptor:  variant,
			ForceVariantSuffix: forceVariantSuffix,
		})
	}

	return nil, results
}
