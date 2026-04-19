package core

import "dryad/task"

func rootSelectorVariantDimensionsByName(dimensions []VariantDimension) map[string]struct{} {
	dimensionsByName := map[string]struct{}{}
	for _, dimension := range dimensions {
		dimensionsByName[dimension.Name] = struct{}{}
	}
	return dimensionsByName
}

func rootSelectorTraitValue(
	ctx *task.ExecutionContext,
	variant *SafeRootVariantReference,
	traitPath string,
) (error, string, bool) {
	if variant.Traits == nil {
		return nil, "", false
	}

	traits := SafeRootTraitsReference{
		BasePath: variant.Traits.BasePath,
		Root:     variant.Root,
	}

	err, trait := traits.Trait(traitPath).Resolve(ctx)
	if err != nil {
		return err, "", false
	}
	if trait == nil {
		return nil, "", false
	}

	err, value := trait.Get(ctx)
	if err != nil {
		return err, "", false
	}

	return nil, value, true
}

func qualifiedSelectorMatchesRootVariantDescriptor(
	ctx *task.ExecutionContext,
	selector qualifiedSelector,
	variant *SafeRootVariantReference,
) (error, bool) {
	if !selector.HasSelector {
		return nil, true
	}

	err, dimensions := variant.Root.VariantDimensions(ctx)
	if err != nil {
		return err, false
	}
	dimensionsByName := rootSelectorVariantDimensionsByName(dimensions)

	for selectorName, requestedOption := range selector.Descriptor {
		if _, isVariantDimension := dimensionsByName[selectorName]; isVariantDimension {
			concreteOption, hasConcreteOption := variant.Descriptor[selectorName]
			err, matches := selectorOptionsMatchVariantValue(
				requestedOption,
				concreteOption,
				hasConcreteOption,
			)
			if err != nil {
				return err, false
			}
			if !matches {
				return nil, false
			}
			continue
		}

		err, traitValue, traitExists := rootSelectorTraitValue(ctx, variant, selectorName)
		if err != nil {
			return err, false
		}

		err, matches := selectorOptionsMatchTraitValue(requestedOption, traitValue, traitExists)
		if err != nil {
			return err, false
		}
		if !matches {
			return nil, false
		}
	}

	return nil, true
}

func qualifiedSelectorMatchesRootVariant(
	ctx *task.ExecutionContext,
	selector qualifiedSelector,
	variant *SafeRootVariantReference,
) (error, bool) {
	path, err := rootVariantGardenPath(variant)
	if err != nil {
		return err, false
	}

	err, pathMatches := qualifiedSelectorPathMatches(selector, path)
	if err != nil {
		return err, false
	}
	if !pathMatches {
		return nil, false
	}

	return qualifiedSelectorMatchesRootVariantDescriptor(ctx, selector, variant)
}

func RootVariantSelectorFilter(request SelectorFilterRequest) (error, RootVariantFilter) {
	err, includeSelectors := parseQualifiedSelectors(request.Include)
	if err != nil {
		return err, nil
	}

	err, excludeSelectors := parseQualifiedSelectors(request.Exclude)
	if err != nil {
		return err, nil
	}

	return nil, func(ctx *task.ExecutionContext, variant *SafeRootVariantReference) (error, bool) {
		matchesInclude := true
		if len(includeSelectors) > 0 {
			err, matchesInclude = qualifiedSelectorsMatchRootVariant(ctx, includeSelectors, variant)
			if err != nil {
				return err, false
			}
		}
		if !matchesInclude {
			return nil, false
		}

		err, matchesExclude := qualifiedSelectorsMatchRootVariant(ctx, excludeSelectors, variant)
		if err != nil {
			return err, false
		}

		return nil, !matchesExclude
	}
}
