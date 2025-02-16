
package core

import (
	"dryad/task"
	// "errors"

	// "path/filepath"

	// "os"

	// zlog "github.com/rs/zerolog/log"
)


func (heapDerivation *UnsafeHeapDerivationReference) Exists(ctx * task.ExecutionContext) (error, bool) {
	var heapDerivationExists bool
	var err error

	heapDerivationExists, err = fileExists(heapDerivation.BasePath)
	return err, heapDerivationExists
}