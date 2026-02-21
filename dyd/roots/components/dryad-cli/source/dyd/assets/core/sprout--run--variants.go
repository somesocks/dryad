package core

import (
	"fmt"
	"os"
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

	if !strings.HasPrefix(name, "stem+") {
		return nil, nil, false
	}

	rawDescriptor := strings.TrimPrefix(name, "stem+")
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

	normalizedSelector := VariantDescriptor{}
	for dimensionName, requestedOption := range selector {
		switch requestedOption {
		case VariantOptionInherit:
			return fmt.Errorf("inherit option is not supported for sprout run variant selectors: %s", dimensionName), nil
		case VariantOptionHost:
			err, hostOption := rootRequirementHostOption(dimensionName)
			if err != nil {
				return err, nil
			}
			normalizedSelector[dimensionName] = hostOption
		default:
			normalizedSelector[dimensionName] = requestedOption
		}
	}

	for dimensionName := range normalizedSelector {
		_, exists := dimensions[dimensionName]
		if !exists {
			return fmt.Errorf("over-specified sprout run variant dimension: %s", dimensionName), nil
		}
	}

	for dimensionName, requestedOption := range normalizedSelector {
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

	matches := make([]sproutRunStemVariant, 0, len(available))
	for _, variant := range available {
		matched := true
		for dimensionName, requestedOption := range normalizedSelector {
			if requestedOption == VariantOptionAny {
				continue
			}

			option, hasDimension := variant.Descriptor[dimensionName]
			if requestedOption == VariantOptionNone {
				if hasDimension {
					matched = false
					break
				}
				continue
			}

			if !hasDimension || option != requestedOption {
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
