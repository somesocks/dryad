

package core

import (
	"dryad/task"
)

type RootFilter func (*task.ExecutionContext, *SafeRootReference) (error, bool)
