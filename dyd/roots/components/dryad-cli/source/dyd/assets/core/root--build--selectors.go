package core

import (
	"dryad/task"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

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

func rootBuild_readAssetsAndCommandsSelectors(
	rootPath string,
) (error, []rootBuildAssetsSelector, []rootBuildCommandsSelector) {
	dydPath := filepath.Join(rootPath, "dyd")
	dydPathExists, err := fileExists(dydPath)
	if err != nil {
		return err, nil, nil
	}
	if !dydPathExists {
		return nil, []rootBuildAssetsSelector{}, []rootBuildCommandsSelector{}
	}

	dydDir, err := os.Open(dydPath)
	if err != nil {
		return err, nil, nil
	}
	defer dydDir.Close()

	assetsSelectors := make([]rootBuildAssetsSelector, 0)
	commandsSelectors := make([]rootBuildCommandsSelector, 0)

	for {
		entryNames, err := dydDir.Readdirnames(64)
		if err != nil && err != io.EOF {
			return err, nil, nil
		}

		for _, selectorName := range entryNames {
			err, isAssetsSelector, assetsDescriptor := rootBuild_parseVariantSelectorDescriptor(
				selectorName,
				"assets",
				"dyd/assets",
			)
			if err != nil {
				return err, nil, nil
			}
			if isAssetsSelector {
				assetsSelectors = append(assetsSelectors, rootBuildAssetsSelector{
					Name:       selectorName,
					Path:       filepath.Join(dydPath, selectorName),
					Descriptor: assetsDescriptor,
				})
			}

			err, isCommandsSelector, commandsDescriptor := rootBuild_parseVariantSelectorDescriptor(
				selectorName,
				"commands",
				"dyd/commands",
			)
			if err != nil {
				return err, nil, nil
			}
			if isCommandsSelector {
				commandsSelectors = append(commandsSelectors, rootBuildCommandsSelector{
					Name:       selectorName,
					Path:       filepath.Join(dydPath, selectorName),
					Descriptor: commandsDescriptor,
				})
			}
		}

		if err == io.EOF {
			break
		}
	}

	return nil, assetsSelectors, commandsSelectors
}

func rootBuild_selectAssetsAndCommandsPaths(
	ctx *task.ExecutionContext,
	rootPath string,
	variantDescriptor string,
) (error, string, string) {
	err, variantContext := RootVariantContextFromFilesystem(variantDescriptor)
	if err != nil {
		return err, "", ""
	}

	rootRef := SafeRootReference{BasePath: rootPath}
	err, dimensions := rootRef.VariantDimensions(ctx)
	if err != nil {
		return err, "", ""
	}

	err, assetsSelectors, commandsSelectors := rootBuild_readAssetsAndCommandsSelectors(rootPath)
	if err != nil {
		return err, "", ""
	}

	err, assetsPath := rootBuild_selectAssetsPathFromSelectors(
		dimensions,
		variantContext.Descriptor,
		variantDescriptor,
		assetsSelectors,
	)
	if err != nil {
		return err, "", ""
	}

	err, commandsPath := rootBuild_selectCommandsPathFromSelectors(
		dimensions,
		variantContext.Descriptor,
		variantDescriptor,
		commandsSelectors,
	)
	if err != nil {
		return err, "", ""
	}

	return nil, assetsPath, commandsPath
}
