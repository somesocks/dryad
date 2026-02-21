package core

import (
	"dryad/task"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

var VARIANT_NAME_RE = regexp.MustCompile(`^[A-Za-z0-9._-]+$`)

var RESERVED_VARIANT_OPTIONS = map[string]struct{}{
	"inherit": {},
	"any":     {},
	"host":    {},
}

type VariantDimensionOption struct {
	Name    string
	Enabled bool
}

type VariantDimension struct {
	Name    string
	Options []VariantDimensionOption
}

func variantNameValid(name string) bool {
	if name == "" {
		return false
	}
	return VARIANT_NAME_RE.MatchString(name)
}

func variantOptionAllowedInCatalog(option string) bool {
	_, isReserved := RESERVED_VARIANT_OPTIONS[option]
	return !isReserved
}

func variantOptionEnabledFromFile(path string) (error, bool) {
	rawBytes, err := os.ReadFile(path)
	if err != nil {
		return err, false
	}

	rawValue := strings.TrimSpace(string(rawBytes))
	switch rawValue {
	case "true":
		return nil, true
	case "false":
		return nil, false
	default:
		return fmt.Errorf("variant option file must contain true or false: %s", path), false
	}
}

func (rootVariants *SafeRootVariantsReference) Dimensions(ctx *task.ExecutionContext) (error, []VariantDimension) {
	dimensionsPath := filepath.Join(rootVariants.BasePath, "dimensions")

	dimensionsExists, err := fileExists(dimensionsPath)
	if err != nil {
		return err, nil
	}

	if !dimensionsExists {
		return nil, []VariantDimension{}
	}

	dimensionsInfo, err := os.Stat(dimensionsPath)
	if err != nil {
		return err, nil
	}
	if !dimensionsInfo.IsDir() {
		return fmt.Errorf("variant dimensions path is not a directory: %s", dimensionsPath), nil
	}

	dimensionEntries, err := os.ReadDir(dimensionsPath)
	if err != nil {
		return err, nil
	}
	sort.Slice(dimensionEntries, func(i int, j int) bool {
		return dimensionEntries[i].Name() < dimensionEntries[j].Name()
	})

	dimensions := make([]VariantDimension, 0, len(dimensionEntries))
	for _, dimensionEntry := range dimensionEntries {
		dimensionName := dimensionEntry.Name()
		if !dimensionEntry.IsDir() {
			return fmt.Errorf("variant dimension must be a directory: %s", filepath.Join(dimensionsPath, dimensionName)), nil
		}

		if !variantNameValid(dimensionName) {
			return fmt.Errorf("invalid variant dimension name: %s", dimensionName), nil
		}

		dimensionPath := filepath.Join(dimensionsPath, dimensionName)
		optionEntries, err := os.ReadDir(dimensionPath)
		if err != nil {
			return err, nil
		}
		sort.Slice(optionEntries, func(i int, j int) bool {
			return optionEntries[i].Name() < optionEntries[j].Name()
		})

		options := make([]VariantDimensionOption, 0, len(optionEntries))
		for _, optionEntry := range optionEntries {
			optionName := optionEntry.Name()
			if optionEntry.IsDir() {
				return fmt.Errorf("variant option must be a file: %s", filepath.Join(dimensionPath, optionName)), nil
			}

			if !variantNameValid(optionName) {
				return fmt.Errorf("invalid variant option name: %s", optionName), nil
			}

			if !variantOptionAllowedInCatalog(optionName) {
				return fmt.Errorf("reserved variant option is not allowed in dimensions: %s", optionName), nil
			}

			optionPath := filepath.Join(dimensionPath, optionName)
			err, optionEnabled := variantOptionEnabledFromFile(optionPath)
			if err != nil {
				return err, nil
			}

			options = append(options, VariantDimensionOption{
				Name:    optionName,
				Enabled: optionEnabled,
			})
		}

		dimensions = append(dimensions, VariantDimension{
			Name:    dimensionName,
			Options: options,
		})
	}

	return nil, dimensions
}
