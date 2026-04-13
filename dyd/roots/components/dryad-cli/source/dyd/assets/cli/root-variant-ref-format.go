package cli

import (
	dryad "dryad/core"
	"dryad/internal/filepath"
)

func formatVariantDescriptorRef(
	basePath string,
	descriptor dryad.VariantDescriptor,
) (error, string) {
	err, variantFilesystem := (dryad.RootVariantContext{Descriptor: descriptor}).Filesystem()
	if err != nil {
		return err, ""
	}

	if variantFilesystem == "" {
		return nil, basePath
	}

	return nil, basePath + dryad.RootRequirementSelectorSeparator + variantFilesystem
}

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

	return formatVariantDescriptorRef(basePath, descriptor)
}
