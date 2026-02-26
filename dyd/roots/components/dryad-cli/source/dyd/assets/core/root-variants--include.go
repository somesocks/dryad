package core

import (
	"dryad/task"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

type VariantInclusion struct {
	Descriptor VariantDescriptor
	Enabled    bool
}

func (rootVariants *SafeRootVariantsReference) Inclusions(ctx *task.ExecutionContext) (error, []VariantInclusion) {
	inclusionsPath := filepath.Join(rootVariants.BasePath, "_include")

	inclusionsExists, err := fileExists(inclusionsPath)
	if err != nil {
		return err, nil
	}

	if !inclusionsExists {
		return nil, []VariantInclusion{}
	}

	inclusionsInfo, err := os.Stat(inclusionsPath)
	if err != nil {
		return err, nil
	}
	if !inclusionsInfo.IsDir() {
		return fmt.Errorf("variant inclusions path is not a directory: %s", inclusionsPath), nil
	}

	inclusionEntries, err := os.ReadDir(inclusionsPath)
	if err != nil {
		return err, nil
	}
	sort.Slice(inclusionEntries, func(i int, j int) bool {
		return inclusionEntries[i].Name() < inclusionEntries[j].Name()
	})

	inclusions := make([]VariantInclusion, 0, len(inclusionEntries))
	for _, inclusionEntry := range inclusionEntries {
		descriptorRaw := inclusionEntry.Name()
		descriptorPath := filepath.Join(inclusionsPath, descriptorRaw)
		if inclusionEntry.IsDir() {
			return fmt.Errorf("included variant descriptor must be a file: %s", descriptorPath), nil
		}

		err, descriptor := variantDescriptorParseFilesystem(descriptorRaw)
		if err != nil {
			return fmt.Errorf("invalid included variant descriptor: %s", descriptorRaw), nil
		}

		err, normalizedDescriptor := variantDescriptorEncodeFilesystem(descriptor)
		if err != nil {
			return err, nil
		}
		if normalizedDescriptor != descriptorRaw {
			return fmt.Errorf("included variant descriptor must be canonical: %s", descriptorRaw), nil
		}

		err, inclusionEnabled := variantOptionEnabledFromFile(descriptorPath)
		if err != nil {
			return err, nil
		}

		inclusions = append(inclusions, VariantInclusion{
			Descriptor: descriptor,
			Enabled:    inclusionEnabled,
		})
	}

	return nil, inclusions
}
