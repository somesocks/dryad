

package core

import (
	"dryad/task"
)

type SproutFilter func (*task.ExecutionContext, *SafeSproutReference) (error, bool)
