package core

import (
	"dryad/task"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func rootBuild_materializeVariantTraits(
	ctx *task.ExecutionContext,
	rootPath string,
	workspacePath string,
	variantDescriptor string,
) error {
	err, variantContext := RootVariantContextFromFilesystem(variantDescriptor)
	if err != nil {
		return err
	}

	rootRef := SafeRootReference{BasePath: rootPath}
	err, dimensions := rootRef.VariantDimensions(ctx)
	if err != nil {
		return err
	}

	traitsSourcePath := filepath.Join(rootPath, "dyd", "traits")
	traitsDestinationPath := filepath.Join(workspacePath, "dyd", "traits")

	sourceTraitsExists, err := fileExists(traitsSourcePath)
	if err != nil {
		return err
	}

	if len(dimensions) == 0 {
		if len(variantContext.Descriptor) > 0 {
			return fmt.Errorf("over-specified root variant descriptor: root has no variant dimensions")
		}
		if sourceTraitsExists {
			err = os.Symlink(traitsSourcePath, traitsDestinationPath)
			if err != nil {
				return err
			}
		}
		return nil
	}

	if sourceTraitsExists {
		err = rootDevelop_copyDir(
			task.SERIAL_CONTEXT,
			traitsSourcePath,
			traitsDestinationPath,
			rootDevelopCopyOptions{ApplyIgnore: false},
		)
		if err != nil {
			return err
		}
	}

	err = os.MkdirAll(traitsDestinationPath, os.ModePerm)
	if err != nil {
		return err
	}

	dimensionsByName := map[string]VariantDimension{}
	for _, dimension := range dimensions {
		dimensionsByName[dimension.Name] = dimension
	}

	for descriptorDimension := range variantContext.Descriptor {
		_, exists := dimensionsByName[descriptorDimension]
		if !exists {
			return fmt.Errorf("over-specified root variant descriptor dimension: %s", descriptorDimension)
		}
	}

	for _, dimension := range dimensions {
		selectedOption, specified := variantContext.Descriptor[dimension.Name]
		if !specified {
			return fmt.Errorf("under-specified root variant descriptor dimension: %s", dimension.Name)
		}

		switch selectedOption {
		case VariantOptionInherit, VariantOptionAny, VariantOptionHost:
			return fmt.Errorf("invalid concrete root variant option for %s: %s", dimension.Name, selectedOption)
		}

		optionByName := map[string]VariantDimensionOption{}
		for _, option := range dimension.Options {
			optionByName[option.Name] = option
		}

		option, exists := optionByName[selectedOption]
		if !exists {
			return fmt.Errorf("wrongly-specified root variant option: %s=%s", dimension.Name, selectedOption)
		}
		if !option.Enabled {
			return fmt.Errorf("disabled root variant option: %s=%s", dimension.Name, selectedOption)
		}

		traitPath := filepath.Join(traitsDestinationPath, dimension.Name)
		traitExists, err := fileExists(traitPath)
		if err != nil {
			return err
		}

		if selectedOption == VariantOptionNone {
			if traitExists {
				return fmt.Errorf("variant descriptor requires omitted trait but trait exists: %s", dimension.Name)
			}
			continue
		}

		if traitExists {
			rawBytes, err := os.ReadFile(traitPath)
			if err != nil {
				return err
			}
			rawValue := strings.TrimSpace(string(rawBytes))
			if rawValue != selectedOption {
				return fmt.Errorf(
					"variant trait conflict for %s: expected %s but found %s",
					dimension.Name,
					selectedOption,
					rawValue,
				)
			}
			continue
		}

		err = os.WriteFile(traitPath, []byte(selectedOption), 0o644)
		if err != nil {
			return err
		}
	}

	workspaceVariantsPath := filepath.Join(traitsDestinationPath, "variants")
	workspaceVariantsExists, err := fileExists(workspaceVariantsPath)
	if err != nil {
		return err
	}
	if workspaceVariantsExists {
		err = os.RemoveAll(workspaceVariantsPath)
		if err != nil {
			return err
		}
	}

	return nil
}
