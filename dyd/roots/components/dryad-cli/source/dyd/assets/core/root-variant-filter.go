package core

import "context"

type RootVariantFilter func(context.Context, *SafeRootVariantReference) (error, bool)

func RootVariantFiltersCompose(filters ...RootVariantFilter) RootVariantFilter {
	return func(ctx context.Context, variant *SafeRootVariantReference) (error, bool) {
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
