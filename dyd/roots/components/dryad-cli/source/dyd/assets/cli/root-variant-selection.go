package cli

import (
	dryad "dryad/core"
	"dryad/task"
	"strings"
)

func resolveSingleRootVariantReference(
	ctx *task.ExecutionContext,
	roots *dryad.SafeRootsReference,
	rootPath string,
	variantRaw string,
) (error, *dryad.SafeRootVariantReference) {
	err, root := roots.Root(rootPath).Resolve(ctx)
	if err != nil {
		return err, nil
	}

	selector := dryad.VariantDescriptor{}
	if strings.TrimSpace(variantRaw) != "" {
		err, variantContext := dryad.RootVariantContextFromFilesystem(variantRaw)
		if err != nil {
			return err, nil
		}
		selector = variantContext.Descriptor
	}

	return root.ResolveBuildVariantReference(
		ctx,
		dryad.RootResolveBuildVariantsRequest{
			Selector: selector,
		},
	)
}
