
package core

import (
	"dryad/task"

	zlog "github.com/rs/zerolog/log"
)


func (ug *UnsafeGardenReference) Resolve(ctx * task.ExecutionContext, _ any) (error, SafeGardenReference) {
	var gardenPath string = ug.BasePath
	var err error

	zlog.Trace().
		Str("BasePath", ug.BasePath).
		Msg("UnsafeGardenReference.Resolve")

	gardenPath, err = GardenPath(ug.BasePath)
	return err, SafeGardenReference{ BasePath: gardenPath }
}