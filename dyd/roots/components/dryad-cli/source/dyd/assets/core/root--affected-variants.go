package core

import (
	"dryad/internal/filepath"
	"dryad/task"
	"strings"
)

func rootAffected_pathWithin(basePath string, path string) (error, bool) {
	if basePath == "" {
		return nil, false
	}

	relPath, err := filepath.Rel(basePath, path)
	if err != nil {
		return err, false
	}

	if relPath == "." {
		return nil, true
	}
	if relPath == ".." || strings.HasPrefix(relPath, ".."+string(filepath.Separator)) {
		return nil, false
	}

	return nil, true
}

func rootAffected_isSelectablePathFamily(relPath string) (error, bool) {
	parts := strings.Split(filepath.Clean(relPath), string(filepath.Separator))
	if len(parts) < 2 || parts[0] != "dyd" {
		return nil, false
	}

	selectorName := parts[1]
	selectorKinds := []struct {
		BaseName string
		Label    string
	}{
		{BaseName: "assets", Label: "dyd/assets"},
		{BaseName: "commands", Label: "dyd/commands"},
		{BaseName: "traits", Label: "dyd/traits"},
		{BaseName: "secrets", Label: "dyd/secrets"},
		{BaseName: "docs", Label: "dyd/docs"},
		{BaseName: "requirements", Label: "dyd/requirements"},
	}

	for _, selectorKind := range selectorKinds {
		err, isSelector, _ := rootBuild_parseVariantSelectorDescriptor(
			selectorName,
			selectorKind.BaseName,
			selectorKind.Label,
		)
		if err != nil {
			return err, false
		}
		if isSelector {
			return nil, true
		}
	}

	return nil, false
}

func rootAffected_pathAffectsAllVariants(rootPath string, changedPath string) (error, bool) {
	relPath, err := filepath.Rel(rootPath, changedPath)
	if err != nil {
		return err, false
	}
	relPath = filepath.Clean(relPath)

	if relPath == "." || relPath == "dyd/type" || relPath == "dyd/variants" || strings.HasPrefix(relPath, filepath.Join("dyd", "variants")+string(filepath.Separator)) {
		return nil, true
	}

	err, isSelectablePathFamily := rootAffected_isSelectablePathFamily(relPath)
	if err != nil {
		return err, false
	}
	if isSelectablePathFamily {
		return nil, false
	}

	return nil, true
}

func rootAffected_requirementsPathMatchesVariant(
	changedPath string,
	requirementsPath string,
	variant VariantDescriptor,
) (error, bool) {
	err, isWithin := rootAffected_pathWithin(requirementsPath, changedPath)
	if err != nil {
		return err, false
	}
	if !isWithin {
		return nil, false
	}

	relPath, err := filepath.Rel(requirementsPath, changedPath)
	if err != nil {
		return err, false
	}
	relPath = filepath.Clean(relPath)

	if relPath == "." || strings.Contains(relPath, string(filepath.Separator)) {
		return nil, true
	}

	err, normalizedName := RootRequirementNormalizeName(relPath)
	if err != nil {
		return nil, true
	}

	err, _, condition := rootRequirementParseName(normalizedName)
	if err != nil {
		return nil, true
	}

	return rootRequirementConditionMatches(variant, condition)
}

func rootAffected_variantMatchesChangedPath(
	changedPath string,
	variant *SafeRootVariantReference,
) (error, bool) {
	selectedPathValues := []string{}
	if variant.Assets != nil {
		selectedPathValues = append(selectedPathValues, variant.Assets.BasePath)
	}
	if variant.Commands != nil {
		selectedPathValues = append(selectedPathValues, variant.Commands.BasePath)
	}
	if variant.Traits != nil {
		selectedPathValues = append(selectedPathValues, variant.Traits.BasePath)
	}
	if variant.Secrets != nil {
		selectedPathValues = append(selectedPathValues, variant.Secrets.BasePath)
	}
	if variant.Docs != nil {
		selectedPathValues = append(selectedPathValues, variant.Docs.BasePath)
	}

	for _, selectedPathValue := range selectedPathValues {
		err, isWithin := rootAffected_pathWithin(selectedPathValue, changedPath)
		if err != nil {
			return err, false
		}
		if isWithin {
			return nil, true
		}
	}

	return rootAffected_requirementsPathMatchesVariant(
		changedPath,
		func() string {
			if variant.Requirements == nil {
				return ""
			}
			return variant.Requirements.BasePath
		}(),
		variant.Descriptor,
	)
}

func (root *SafeRootReference) ResolveAffectedVariants(
	ctx *task.ExecutionContext,
	changedPaths []string,
) (error, []VariantDescriptor) {
	err, resolvedVariants := root.ResolveBuildVariantReferences(
		ctx,
		RootResolveBuildVariantsRequest{},
	)
	if err != nil {
		return err, nil
	}

	for _, changedPath := range changedPaths {
		err, affectsAllVariants := rootAffected_pathAffectsAllVariants(root.BasePath, changedPath)
		if err != nil {
			return err, nil
		}
		if affectsAllVariants {
			results := make([]VariantDescriptor, 0, len(resolvedVariants))
			for _, resolvedVariant := range resolvedVariants {
				results = append(results, resolvedVariant.Descriptor)
			}
			return nil, results
		}
	}

	affectedVariants := make([]VariantDescriptor, 0, len(resolvedVariants))

	for _, resolvedVariant := range resolvedVariants {
		isAffected := false
		for _, changedPath := range changedPaths {
			err, matchesVariant := rootAffected_variantMatchesChangedPath(
				changedPath,
				resolvedVariant,
			)
			if err != nil {
				return err, nil
			}
			if matchesVariant {
				isAffected = true
				break
			}
		}

		if isAffected {
			affectedVariants = append(affectedVariants, resolvedVariant.Descriptor)
		}
	}

	return nil, affectedVariants
}
