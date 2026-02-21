package core

import (
	"fmt"
	"strings"

	"dryad/task"
)

func rootDevelop_resolveVariant(
	ctx *task.ExecutionContext,
	root *SafeRootReference,
	rawSelector string,
) (error, string) {
	err, selectorDescriptor := normalizeRootBuildVariantDescriptor(rawSelector)
	if err != nil {
		return err, ""
	}

	err, selector := variantDescriptorParseFilesystem(selectorDescriptor)
	if err != nil {
		return err, ""
	}

	err, variants := root.ResolveBuildVariants(
		ctx,
		RootResolveBuildVariantsRequest{
			Selector:                selector,
			IgnoreUnknownDimensions: true,
		},
	)
	if err != nil {
		return err, ""
	}

	if len(variants) == 0 {
		return fmt.Errorf("no root develop variants resolved"), ""
	}

	if len(variants) > 1 {
		rendered := make([]string, 0, len(variants))
		for _, variant := range variants {
			err, descriptor := variantDescriptorEncodeFilesystem(variant)
			if err != nil {
				return err, ""
			}
			rendered = append(rendered, descriptor)
		}
		return fmt.Errorf(
			"ambiguous root develop variant selector: resolved %d variants (%s)",
			len(variants),
			strings.Join(rendered, ", "),
		), ""
	}

	err, concreteDescriptor := variantDescriptorEncodeFilesystem(variants[0])
	if err != nil {
		return err, ""
	}

	return nil, concreteDescriptor
}
