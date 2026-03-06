package core

import (
	"dryad/internal/os"
	"dryad/task"
	"fmt"
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

func normalizeVariantFilterDescriptor(descriptor VariantDescriptor) VariantDescriptor {
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

func rootVariantFilterResolveChoicesForDimension(
	dimension VariantDimension,
	requestedOptionRaw string,
	filterKind string,
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
			return fmt.Errorf("wrongly-specified %s variant option: %s=%s", filterKind, dimension.Name, optionName), rootVariantExclusionOptionChoice{}
		}
		if !enabled[optionName] {
			return fmt.Errorf("disabled %s variant option: %s=%s", filterKind, dimension.Name, optionName), rootVariantExclusionOptionChoice{}
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
				return fmt.Errorf("no enabled variant options for %s variant resolution: %s", filterKind, dimension.Name), nil
			}
		case VariantOptionInherit:
			return fmt.Errorf("inherit option is not supported for %s variant selectors: %s", filterKind, dimension.Name), nil
		case VariantOptionHost:
			return fmt.Errorf("host option is not supported for %s variant selectors: %s", filterKind, dimension.Name), nil
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

type variantRule struct {
	Descriptor VariantDescriptor
	Enabled    bool
}

func expandVariantRule(
	dimensionsByName map[string]VariantDimension,
	dimensionNames []string,
	rule variantRule,
	filterKind string,
) (error, []VariantDescriptor) {
	for descriptorDimension := range rule.Descriptor {
		_, exists := dimensionsByName[descriptorDimension]
		if !exists {
			return fmt.Errorf("over-specified %s variant dimension: %s", filterKind, descriptorDimension), nil
		}
	}

	resolvedExclusions := []VariantDescriptor{{}}
	for _, dimensionName := range dimensionNames {
		dimension := dimensionsByName[dimensionName]

		requestedOption, hasRequestedOption := rule.Descriptor[dimensionName]
		if !hasRequestedOption {
			requestedOption = VariantOptionAny
		}

		err, choices := rootVariantFilterResolveChoicesForDimension(
			dimension,
			requestedOption,
			filterKind,
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

func variantRulesFromExclusions(exclusions []VariantExclusion) []variantRule {
	rules := make([]variantRule, 0, len(exclusions))
	for _, exclusion := range exclusions {
		rules = append(rules, variantRule{
			Descriptor: exclusion.Descriptor,
			Enabled:    exclusion.Enabled,
		})
	}
	return rules
}

func variantRulesFromInclusions(inclusions []VariantInclusion) []variantRule {
	rules := make([]variantRule, 0, len(inclusions))
	for _, inclusion := range inclusions {
		rules = append(rules, variantRule{
			Descriptor: inclusion.Descriptor,
			Enabled:    inclusion.Enabled,
		})
	}
	return rules
}

func expandVariantRulesToMap(
	dimensions []VariantDimension,
	rules []variantRule,
	filterKind string,
) (error, map[string]struct{}) {
	dimensionsByName := map[string]VariantDimension{}
	dimensionNames := make([]string, 0, len(dimensions))
	for _, dimension := range dimensions {
		dimensionsByName[dimension.Name] = dimension
		dimensionNames = append(dimensionNames, dimension.Name)
	}
	sort.Strings(dimensionNames)

	filteredVariants := map[string]struct{}{}
	for _, rule := range rules {
		err, expandedRules := expandVariantRule(dimensionsByName, dimensionNames, rule, filterKind)
		if err != nil {
			return err, nil
		}

		if !rule.Enabled {
			continue
		}

		for _, expandedRule := range expandedRules {
			err, ruleKey := variantDescriptorEncodeFilesystem(
				normalizeVariantFilterDescriptor(expandedRule),
			)
			if err != nil {
				return err, nil
			}
			filteredVariants[ruleKey] = struct{}{}
		}
	}

	return nil, filteredVariants
}

func applyVariantFilters(
	dimensions []VariantDimension,
	inclusions []VariantInclusion,
	exclusions []VariantExclusion,
	variants []VariantDescriptor,
) (error, []VariantDescriptor) {
	err, includedVariants := expandVariantRulesToMap(dimensions, variantRulesFromInclusions(inclusions), "included")
	if err != nil {
		return err, nil
	}
	err, excludedVariants := expandVariantRulesToMap(dimensions, variantRulesFromExclusions(exclusions), "excluded")
	if err != nil {
		return err, nil
	}

	if len(includedVariants) == 0 && len(excludedVariants) == 0 {
		return nil, variants
	}

	filteredVariants := make([]VariantDescriptor, 0, len(variants))
	for _, variant := range variants {
		err, variantKey := variantDescriptorEncodeFilesystem(variant)
		if err != nil {
			return err, nil
		}
		if len(includedVariants) > 0 {
			_, isIncluded := includedVariants[variantKey]
			if !isIncluded {
				continue
			}
		}
		_, isExcluded := excludedVariants[variantKey]
		if isExcluded {
			continue
		}
		filteredVariants = append(filteredVariants, variant)
	}

	return nil, filteredVariants
}
