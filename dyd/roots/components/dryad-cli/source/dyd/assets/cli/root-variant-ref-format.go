package cli

import (
	dryad "dryad/core"
	"dryad/internal/filepath"
)

func formatRootVariantRef(
	variant *dryad.SafeRootVariantReference,
	relative bool,
) (error, string) {
	basePath := variant.Root.BasePath
	if relative {
		var err error
		basePath, err = filepath.Rel(
			variant.Root.Roots.Garden.BasePath,
			variant.Root.BasePath,
		)
		if err != nil {
			return err, ""
		}
	}

	err, variantURL := variant.URL()
	if err != nil {
		return err, ""
	}

	return nil, basePath + variantURL
}
