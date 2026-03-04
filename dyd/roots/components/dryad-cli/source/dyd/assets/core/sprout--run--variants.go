package core

import (
	"dryad/internal/os"
	"fmt"
	"path/filepath"
	"sort"
	"strings"
)

type sproutRunStemVariant struct {
	Descriptor    VariantDescriptor
	DescriptorRaw string
	StemPath      string
}

func sproutRunStemVariantDependencyDescriptor(name string) (error, VariantDescriptor, bool) {
	if name == "stem" {
		return nil, VariantDescriptor{}, true
	}

	stemPrefix := "stem" + RootRequirementSelectorSeparator
	if !strings.HasPrefix(name, stemPrefix) {
		return nil, nil, false
	}

	rawDescriptor := strings.TrimPrefix(name, stemPrefix)
	err, descriptor := variantDescriptorParseFilesystem(rawDescriptor)
	if err != nil {
		return fmt.Errorf("invalid sprout stem dependency descriptor: %s", name), nil, false
	}

	err, normalizedDescriptor := variantDescriptorEncodeFilesystem(descriptor)
	if err != nil {
		return err, nil, false
	}
	if normalizedDescriptor != rawDescriptor {
		return fmt.Errorf("non-canonical sprout stem dependency descriptor: %s", name), nil, false
	}

	return nil, descriptor, true
}

func (sprout *SafeSproutReference) runStemVariants() (error, []sproutRunStemVariant) {
	dependenciesPath := filepath.Join(sprout.BasePath, "dyd", "dependencies")
	dependencyEntries, err := os.ReadDir(dependenciesPath)
	if err != nil {
		return err, nil
	}

	byDescriptor := map[string]sproutRunStemVariant{}
	descriptors := []string{}
	for _, dependencyEntry := range dependencyEntries {
		dependencyName := dependencyEntry.Name()

		err, descriptor, isStemDependency := sproutRunStemVariantDependencyDescriptor(dependencyName)
		if err != nil {
			return err, nil
		}
		if !isStemDependency {
			continue
		}

		dependencyPath := filepath.Join(dependenciesPath, dependencyName)
		resolvedDependencyPath, err := filepath.EvalSymlinks(dependencyPath)
		if err != nil {
			return err, nil
		}

		stemPath, err := StemPath(resolvedDependencyPath)
		if err != nil {
			return err, nil
		}

		err, descriptorRaw := variantDescriptorEncodeFilesystem(descriptor)
		if err != nil {
			return err, nil
		}

		if _, exists := byDescriptor[descriptorRaw]; exists {
			return fmt.Errorf("duplicate sprout stem dependency for variant descriptor: %s", descriptorRaw), nil
		}

		byDescriptor[descriptorRaw] = sproutRunStemVariant{
			Descriptor:    descriptor,
			DescriptorRaw: descriptorRaw,
			StemPath:      stemPath,
		}
		descriptors = append(descriptors, descriptorRaw)
	}

	if len(descriptors) == 0 {
		return fmt.Errorf("sprout has no stem dependencies: %s", sprout.BasePath), nil
	}

	sort.Strings(descriptors)

	variants := make([]sproutRunStemVariant, 0, len(descriptors))
	for _, descriptor := range descriptors {
		variants = append(variants, byDescriptor[descriptor])
	}

	return nil, variants
}

func resolveSproutRunVariantSelector(raw string) (error, VariantDescriptor) {
	err, variantContext := RootVariantContextFromFilesystem(raw)
	if err != nil {
		return err, nil
	}
	return normalizeVariantDescriptor(variantContext.Descriptor)
}

func resolveSproutRunStemVariants(
	available []sproutRunStemVariant,
	selector VariantDescriptor,
) (error, []sproutRunStemVariant) {
	dimensions := map[string]struct{}{}
	for _, variant := range available {
		for dimensionName := range variant.Descriptor {
			dimensions[dimensionName] = struct{}{}
		}
	}

	normalizedSelector := map[string][]string{}
	for dimensionName, requestedOptionRaw := range selector {
		err, requestedOptions := variantDescriptorOptionValues(requestedOptionRaw)
		if err != nil {
			return err, nil
		}

		normalizedOptions := make([]string, 0, len(requestedOptions))
		seenOptions := map[string]struct{}{}
		for _, requestedOption := range requestedOptions {
			normalizedOption := requestedOption

			switch requestedOption {
			case VariantOptionInherit:
				return fmt.Errorf("inherit option is not supported for sprout run variant selectors: %s", dimensionName), nil
			case VariantOptionHost:
				err, hostOption := rootRequirementHostOption(dimensionName)
				if err != nil {
					return err, nil
				}
				normalizedOption = hostOption
			}

			if _, exists := seenOptions[normalizedOption]; exists {
				continue
			}
			seenOptions[normalizedOption] = struct{}{}
			normalizedOptions = append(normalizedOptions, normalizedOption)
		}

		normalizedSelector[dimensionName] = normalizedOptions
	}

	for dimensionName := range normalizedSelector {
		_, exists := dimensions[dimensionName]
		if !exists {
			return fmt.Errorf("over-specified sprout run variant dimension: %s", dimensionName), nil
		}
	}

	for dimensionName, requestedOptions := range normalizedSelector {
		for _, requestedOption := range requestedOptions {
			if requestedOption == VariantOptionAny {
				continue
			}

			optionExists := false
			for _, variant := range available {
				option, hasDimension := variant.Descriptor[dimensionName]
				if requestedOption == VariantOptionNone {
					if !hasDimension {
						optionExists = true
						break
					}
					continue
				}

				if hasDimension && option == requestedOption {
					optionExists = true
					break
				}
			}

			if !optionExists {
				return fmt.Errorf("wrongly-specified sprout run variant option: %s=%s", dimensionName, requestedOption), nil
			}
		}
	}

	matches := make([]sproutRunStemVariant, 0, len(available))
	for _, variant := range available {
		matched := true
		for dimensionName, requestedOptions := range normalizedSelector {
			option, hasDimension := variant.Descriptor[dimensionName]
			dimensionMatched := false
			for _, requestedOption := range requestedOptions {
				if requestedOption == VariantOptionAny {
					dimensionMatched = true
					break
				}

				if requestedOption == VariantOptionNone {
					if !hasDimension {
						dimensionMatched = true
						break
					}
					continue
				}

				if hasDimension && option == requestedOption {
					dimensionMatched = true
					break
				}
			}

			if !dimensionMatched {
				matched = false
				break
			}
		}

		if matched {
			matches = append(matches, variant)
		}
	}

	if len(matches) == 0 {
		return fmt.Errorf("resolved sprout run variants are empty"), nil
	}

	sort.Slice(matches, func(i int, j int) bool {
		return matches[i].DescriptorRaw < matches[j].DescriptorRaw
	})

	return nil, matches
}
