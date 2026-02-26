package core

import (
	"dryad/task"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type VariantExclusion struct {
	Descriptor VariantDescriptor
	Enabled    bool
}

func (rootVariants *SafeRootVariantsReference) Exclusions(ctx *task.ExecutionContext) (error, []VariantExclusion) {
	exclusionsPath := filepath.Join(rootVariants.BasePath, "_exclude")

	exclusionsExists, err := fileExists(exclusionsPath)
	if err != nil {
		return err, nil
	}

	if !exclusionsExists {
		return nil, []VariantExclusion{}
	}

	exclusionsInfo, err := os.Stat(exclusionsPath)
	if err != nil {
		return err, nil
	}
	if !exclusionsInfo.IsDir() {
		return fmt.Errorf("variant exclusions path is not a directory: %s", exclusionsPath), nil
	}

	exclusionEntries, err := os.ReadDir(exclusionsPath)
	if err != nil {
		return err, nil
	}
	sort.Slice(exclusionEntries, func(i int, j int) bool {
		return exclusionEntries[i].Name() < exclusionEntries[j].Name()
	})

	exclusions := make([]VariantExclusion, 0, len(exclusionEntries))
	for _, exclusionEntry := range exclusionEntries {
		descriptorRaw := exclusionEntry.Name()
		descriptorPath := filepath.Join(exclusionsPath, descriptorRaw)
		if exclusionEntry.IsDir() {
			return fmt.Errorf("excluded variant descriptor must be a file: %s", descriptorPath), nil
		}

		err, descriptor := variantDescriptorParseFilesystem(descriptorRaw)
		if err != nil {
			return fmt.Errorf("invalid excluded variant descriptor: %s", descriptorRaw), nil
		}

		err, normalizedDescriptor := variantDescriptorEncodeFilesystem(descriptor)
		if err != nil {
			return err, nil
		}
		if normalizedDescriptor != descriptorRaw {
			return fmt.Errorf("excluded variant descriptor must be canonical: %s", descriptorRaw), nil
		}

		err, exclusionEnabled := variantOptionEnabledFromFile(descriptorPath)
		if err != nil {
			return err, nil
		}

		exclusions = append(exclusions, VariantExclusion{
			Descriptor: descriptor,
			Enabled:    exclusionEnabled,
		})
	}

	return nil, exclusions
}

func normalizeVariantExclusionDescriptor(descriptor VariantDescriptor) VariantDescriptor {
	normalized := VariantDescriptor{}
	for dimensionName, optionName := range descriptor {
		if optionName == VariantOptionNone {
			continue
		}
		normalized[dimensionName] = optionName
	}
	return normalized
}

type rootVariantExclusionOptionChoice struct {
	Omit   bool
	Option string
}

func rootVariantExclusionResolveChoicesForDimension(
	dimension VariantDimension,
	requestedOptionRaw string,
) (error, []rootVariantExclusionOptionChoice) {
	exists := map[string]bool{}
	enabled := map[string]bool{}
	choices := make([]rootVariantExclusionOptionChoice, 0)
	seenChoices := map[string]struct{}{}

	for _, option := range dimension.Options {
		exists[option.Name] = true
		enabled[option.Name] = option.Enabled
	}

	requireEnabledOption := func(optionName string) (error, rootVariantExclusionOptionChoice) {
		if !exists[optionName] {
			return fmt.Errorf("wrongly-specified excluded variant option: %s=%s", dimension.Name, optionName), rootVariantExclusionOptionChoice{}
		}
		if !enabled[optionName] {
			return fmt.Errorf("disabled excluded variant option: %s=%s", dimension.Name, optionName), rootVariantExclusionOptionChoice{}
		}

		if optionName == VariantOptionNone {
			return nil, rootVariantExclusionOptionChoice{Omit: true}
		}

		return nil, rootVariantExclusionOptionChoice{Option: optionName}
	}

	appendUniqueChoice := func(choice rootVariantExclusionOptionChoice) {
		key := choice.Option
		if choice.Omit {
			key = VariantOptionNone
		}
		if _, alreadyIncluded := seenChoices[key]; alreadyIncluded {
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
					appendUniqueChoice(rootVariantExclusionOptionChoice{Omit: true})
					continue
				}
				appendUniqueChoice(rootVariantExclusionOptionChoice{Option: option.Name})
			}
			if len(choices) == 0 {
				return fmt.Errorf("no enabled variant options for excluded variant resolution: %s", dimension.Name), nil
			}
		case VariantOptionInherit:
			return fmt.Errorf("inherit option is not supported for excluded variant selectors: %s", dimension.Name), nil
		case VariantOptionHost:
			return fmt.Errorf("host option is not supported for excluded variant selectors: %s", dimension.Name), nil
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

func expandVariantExclusion(
	dimensionsByName map[string]VariantDimension,
	dimensionNames []string,
	exclusion VariantExclusion,
) (error, []VariantDescriptor) {
	for descriptorDimension := range exclusion.Descriptor {
		_, exists := dimensionsByName[descriptorDimension]
		if !exists {
			return fmt.Errorf("over-specified excluded variant dimension: %s", descriptorDimension), nil
		}
	}

	resolvedExclusions := []VariantDescriptor{{}}
	for _, dimensionName := range dimensionNames {
		dimension := dimensionsByName[dimensionName]

		requestedOption, hasRequestedOption := exclusion.Descriptor[dimensionName]
		if !hasRequestedOption {
			return fmt.Errorf("under-specified excluded variant dimension: %s", dimensionName), nil
		}

		err, choices := rootVariantExclusionResolveChoicesForDimension(
			dimension,
			requestedOption,
		)
		if err != nil {
			return err, nil
		}

		nextResolvedExclusions := make([]VariantDescriptor, 0, len(resolvedExclusions)*len(choices))
		for _, baseDescriptor := range resolvedExclusions {
			for _, choice := range choices {
				nextDescriptor := cloneVariantDescriptor(baseDescriptor)
				if !choice.Omit {
					nextDescriptor[dimensionName] = choice.Option
				}
				nextResolvedExclusions = append(nextResolvedExclusions, nextDescriptor)
			}
		}
		resolvedExclusions = nextResolvedExclusions
	}

	return nil, resolvedExclusions
}

func applyVariantExclusions(
	dimensions []VariantDimension,
	exclusions []VariantExclusion,
	variants []VariantDescriptor,
) (error, []VariantDescriptor) {
	dimensionsByName := map[string]VariantDimension{}
	dimensionNames := make([]string, 0, len(dimensions))
	for _, dimension := range dimensions {
		dimensionsByName[dimension.Name] = dimension
		dimensionNames = append(dimensionNames, dimension.Name)
	}
	sort.Strings(dimensionNames)

	excludedVariants := map[string]struct{}{}
	for _, exclusion := range exclusions {
		err, expandedExclusions := expandVariantExclusion(dimensionsByName, dimensionNames, exclusion)
		if err != nil {
			return err, nil
		}

		if !exclusion.Enabled {
			continue
		}

		for _, expandedExclusion := range expandedExclusions {
			err, exclusionKey := variantDescriptorEncodeFilesystem(
				normalizeVariantExclusionDescriptor(expandedExclusion),
			)
			if err != nil {
				return err, nil
			}
			excludedVariants[exclusionKey] = struct{}{}
		}
	}

	if len(excludedVariants) == 0 {
		return nil, variants
	}

	filteredVariants := make([]VariantDescriptor, 0, len(variants))
	for _, variant := range variants {
		err, variantKey := variantDescriptorEncodeFilesystem(variant)
		if err != nil {
			return err, nil
		}
		_, isExcluded := excludedVariants[variantKey]
		if isExcluded {
			continue
		}
		filteredVariants = append(filteredVariants, variant)
	}

	return nil, filteredVariants
}
