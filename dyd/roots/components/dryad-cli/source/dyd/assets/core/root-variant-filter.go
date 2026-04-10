package core

import "dryad/task"

type RootVariantFilter func(*task.ExecutionContext, *SafeRootVariantReference) (error, bool)

func RootVariantFiltersCompose(filters ...RootVariantFilter) RootVariantFilter {
	return func(ctx *task.ExecutionContext, variant *SafeRootVariantReference) (error, bool) {
		for _, filter := range filters {
			err, match := filter(ctx, variant)
			if err != nil {
				return err, false
			} else if !match {
				return nil, false
			}
		}

		return nil, true
	}
}

func RootVariantFilterToRootFilterAny(filter RootVariantFilter) RootFilter {
	return func(ctx *task.ExecutionContext, root *SafeRootReference) (error, bool) {
		err, variants := root.ResolveBuildVariantReferences(
			ctx,
			RootResolveBuildVariantsRequest{},
		)
		if err != nil {
			return err, false
		}

		for _, variant := range variants {
			err, match := filter(ctx, variant)
			if err != nil {
				return err, false
			}
			if match {
				return nil, true
			}
		}

		return nil, false
	}
}
