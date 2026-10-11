package core

import "context"

func SproutFiltersCompose(filters ...SproutFilter) SproutFilter {
	return func(ctx context.Context, sprout *SafeSproutReference) (error, bool) {
		for _, filter := range filters {
			err, match := filter(ctx, sprout)
			if err != nil {
				return err, false
			} else if !match {
				return nil, false
			}
		}

		return nil, true
	}
}
