package cli

import (
	dryad "dryad/core"
	"dryad/internal/filepath"
)

func formatRootVariantRef(
	variant *dryad.SafeRootVariantReference,
	relative bool,
) (error, string) {
	return formatRootVariantDescriptorRef(variant.Root, variant.Descriptor, relative)
}

func formatRootVariantDescriptorRef(
	root *dryad.SafeRootReference,
	descriptor dryad.VariantDescriptor,
	relative bool,
) (error, string) {
	basePath := root.BasePath
	if relative {
		var err error
		basePath, err = filepath.Rel(
			root.Roots.Garden.BasePath,
			root.BasePath,
		)
		if err != nil {
			return err, ""
		}
	}

	err, variantURL := (dryad.RootVariantContext{Descriptor: descriptor}).URL()
	if err != nil {
		return err, ""
	}

	return nil, basePath + variantURL
}
