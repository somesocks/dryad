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
	exclusionsPath := filepath.Join(rootVariants.BasePath, "exclude")

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

func applyVariantExclusions(
	dimensions []VariantDimension,
	exclusions []VariantExclusion,
	variants []VariantDescriptor,
) (error, []VariantDescriptor) {
	optionStateByDimension := map[string]map[string]bool{}
	for _, dimension := range dimensions {
		optionState := map[string]bool{}
		for _, option := range dimension.Options {
			optionState[option.Name] = option.Enabled
		}
		optionStateByDimension[dimension.Name] = optionState
	}

	excludedVariants := map[string]struct{}{}
	for _, exclusion := range exclusions {
		for descriptorDimension, descriptorOption := range exclusion.Descriptor {
			optionState, exists := optionStateByDimension[descriptorDimension]
			if !exists {
				return fmt.Errorf("over-specified excluded variant dimension: %s", descriptorDimension), nil
			}

			optionEnabled, exists := optionState[descriptorOption]
			if !exists {
				return fmt.Errorf("wrongly-specified excluded variant option: %s=%s", descriptorDimension, descriptorOption), nil
			}
			if !optionEnabled {
				return fmt.Errorf("disabled excluded variant option: %s=%s", descriptorDimension, descriptorOption), nil
			}
			if descriptorOption == VariantOptionAny || descriptorOption == VariantOptionInherit || descriptorOption == VariantOptionHost {
				return fmt.Errorf("invalid excluded variant option: %s=%s", descriptorDimension, descriptorOption), nil
			}
		}

		for _, dimension := range dimensions {
			_, hasOption := exclusion.Descriptor[dimension.Name]
			if !hasOption {
				return fmt.Errorf("under-specified excluded variant dimension: %s", dimension.Name), nil
			}
		}

		if !exclusion.Enabled {
			continue
		}

		err, exclusionKey := variantDescriptorEncodeFilesystem(
			normalizeVariantExclusionDescriptor(exclusion.Descriptor),
		)
		if err != nil {
			return err, nil
		}
		excludedVariants[exclusionKey] = struct{}{}
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
