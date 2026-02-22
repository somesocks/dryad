package core

import (
	"dryad/task"
	"fmt"
	"sort"
)

type RootResolveBuildVariantsRequest struct {
	Selector                VariantDescriptor
	IgnoreUnknownDimensions bool
}

type rootBuildOptionChoice struct {
	Omit   bool
	Option string
}

func rootBuildResolveChoicesForDimension(
	dimension VariantDimension,
	requestedOptionRaw string,
) (error, []rootBuildOptionChoice) {
	exists := map[string]bool{}
	enabled := map[string]bool{}
	choices := make([]rootBuildOptionChoice, 0)
	seenChoices := map[string]struct{}{}

	for _, option := range dimension.Options {
		exists[option.Name] = true
		enabled[option.Name] = option.Enabled
	}

	requireEnabledOption := func(optionName string) (error, rootBuildOptionChoice) {
		if !exists[optionName] {
			return fmt.Errorf("wrongly-specified root build variant option: %s=%s", dimension.Name, optionName), rootBuildOptionChoice{}
		}
		if !enabled[optionName] {
			return fmt.Errorf("disabled root build variant option: %s=%s", dimension.Name, optionName), rootBuildOptionChoice{}
		}

		if optionName == VariantOptionNone {
			return nil, rootBuildOptionChoice{Omit: true}
		}

		return nil, rootBuildOptionChoice{Option: optionName}
	}

	appendUniqueChoice := func(choice rootBuildOptionChoice) {
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
		case VariantOptionAny:
			for _, option := range dimension.Options {
				if !option.Enabled {
					continue
				}

				if option.Name == VariantOptionNone {
					appendUniqueChoice(rootBuildOptionChoice{Omit: true})
					continue
				}

				appendUniqueChoice(rootBuildOptionChoice{Option: option.Name})
			}

			if len(choices) == 0 {
				return fmt.Errorf("no enabled variant options for root build resolution: %s", dimension.Name), nil
			}

		case VariantOptionInherit:
			return fmt.Errorf("inherit option is not supported for root build variant selectors: %s", dimension.Name), nil

		case VariantOptionHost:
			return fmt.Errorf("host option is not supported for root build variant selectors: %s", dimension.Name), nil

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

func (root *SafeRootReference) ResolveBuildVariants(
	ctx *task.ExecutionContext,
	req RootResolveBuildVariantsRequest,
) (error, []VariantDescriptor) {
	err, selector := normalizeVariantDescriptor(req.Selector)
	if err != nil {
		return err, nil
	}

	err, dimensions := root.VariantDimensions(ctx)
	if err != nil {
		return err, nil
	}
	err, exclusions := root.VariantExclusions(ctx)
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

	effectiveSelector := VariantDescriptor{}
	for selectorDimension, selectorOption := range selector {
		_, exists := dimensionsByName[selectorDimension]
		if !exists {
			if req.IgnoreUnknownDimensions {
				continue
			}
			return fmt.Errorf("over-specified root build variant dimension: %s", selectorDimension), nil
		}

		effectiveSelector[selectorDimension] = selectorOption
	}

	resolvedVariants := []VariantDescriptor{{}}
	for _, dimensionName := range dimensionNames {
		dimension := dimensionsByName[dimensionName]

		requestedOption, hasRequestedOption := effectiveSelector[dimensionName]
		if !hasRequestedOption {
			requestedOption = VariantOptionAny
		}

		err, choices := rootBuildResolveChoicesForDimension(
			dimension,
			requestedOption,
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

	err, results = applyVariantExclusions(dimensions, exclusions, results)
	if err != nil {
		return err, nil
	}
	if len(results) == 0 {
		return fmt.Errorf("resolved root build variants are excluded by variants/_exclude"), nil
	}

	return nil, results
}
