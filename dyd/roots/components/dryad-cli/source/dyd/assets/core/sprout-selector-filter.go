package core

import (
	"dryad/task"
	"errors"
)

func sproutSelectorHasVariantDimension(variants []sproutRunStemVariant, dimensionName string) bool {
	for _, variant := range variants {
		if _, exists := variant.Descriptor[dimensionName]; exists {
			return true
		}
	}

	return false
}

func sproutSelectorVariantsMatch(
	selectorName string,
	requestedOption string,
	variants []sproutRunStemVariant,
) (error, bool) {
	for _, variant := range variants {
		concreteOption, hasConcreteOption := variant.Descriptor[selectorName]
		err, matches := selectorOptionsMatchVariantValue(
			requestedOption,
			concreteOption,
			hasConcreteOption,
		)
		if err != nil {
			return err, false
		}
		if matches {
			return nil, true
		}
	}

	return nil, false
}

func sproutSelectorTraitValue(
	ctx *task.ExecutionContext,
	sprout *SafeSproutReference,
	traitPath string,
) (error, string, bool) {
	err, traits := sprout.Traits().Resolve(ctx)
	if err != nil {
		if errors.Is(err, ErrorNoSproutTraits) {
			return nil, "", false
		}
		return err, "", false
	}

	err, trait := traits.Trait(traitPath).Resolve(ctx)
	if err != nil {
		if errors.Is(err, ErrorNoSproutTrait) {
			return nil, "", false
		}
		return err, "", false
	}

	err, value := trait.Get(ctx)
	if err != nil {
		return err, "", false
	}

	return nil, value, true
}

func qualifiedSelectorMatchesSproutDescriptor(
	ctx *task.ExecutionContext,
	selector qualifiedSelector,
	sprout *SafeSproutReference,
) (error, bool) {
	if !selector.HasSelector {
		return nil, true
	}

	err, variants := sprout.runStemVariants()
	if err != nil {
		return err, false
	}

	for selectorName, requestedOption := range selector.Descriptor {
		if sproutSelectorHasVariantDimension(variants, selectorName) {
			err, matches := sproutSelectorVariantsMatch(selectorName, requestedOption, variants)
			if err != nil {
				return err, false
			}
			if !matches {
				return nil, false
			}
			continue
		}

		err, traitValue, traitExists := sproutSelectorTraitValue(ctx, sprout, selectorName)
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

func qualifiedSelectorMatchesSprout(
	ctx *task.ExecutionContext,
	selector qualifiedSelector,
	sprout *SafeSproutReference,
) (error, bool) {
	path, err := sproutGardenPath(sprout)
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

	return qualifiedSelectorMatchesSproutDescriptor(ctx, selector, sprout)
}

func SproutSelectorFilter(request SelectorFilterRequest) (error, SproutFilter) {
	err, includeSelectors := parseQualifiedSelectors(request.Include)
	if err != nil {
		return err, nil
	}

	err, excludeSelectors := parseQualifiedSelectors(request.Exclude)
	if err != nil {
		return err, nil
	}

	return nil, func(ctx *task.ExecutionContext, sprout *SafeSproutReference) (error, bool) {
		matchesInclude := true
		if len(includeSelectors) > 0 {
			err, matchesInclude = qualifiedSelectorsMatchSprout(ctx, includeSelectors, sprout)
			if err != nil {
				return err, false
			}
		}
		if !matchesInclude {
			return nil, false
		}

		err, matchesExclude := qualifiedSelectorsMatchSprout(ctx, excludeSelectors, sprout)
		if err != nil {
			return err, false
		}

		return nil, !matchesExclude
	}
}
