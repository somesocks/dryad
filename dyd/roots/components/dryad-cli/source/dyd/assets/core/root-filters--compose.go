

package core

import (
	"dryad/task"
)

func RootFiltersCompose (filters ...RootFilter) RootFilter {
	return func (ctx *task.ExecutionContext, root *SafeRootReference) (error, bool) {
		for _, filter := range filters {
			err, match := filter(ctx, root)
			if err != nil {
				return err, false
			} else if !match {
				return nil, false
			}
		}

		return nil, true
	}
}