
package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	// "os"

	zlog "github.com/rs/zerolog/log"
)


func (sprouts *UnsafeSproutsReference) Resolve(ctx *task.ExecutionContext) (error, *SafeSproutsReference) {
	zlog.Trace().
		Str("path", sprouts.BasePath).
		Msg("UnsafeSproutsReference.Resolve")

	var sproutsExists bool
	var err error
	var safeRef SafeSproutsReference

	sproutsExists, err = fileExists(sprouts.BasePath)
	if err != nil {
		return err, nil
	}

	if !sproutsExists {
		// err := os.Mkdir(sprouts.BasePath, os.ModePerm)
		err, _ := fs2.Mkdir2(
			ctx,
			fs2.MkdirRequest{
				Path: sprouts.BasePath,
				Mode: 0o551,
			},
		)
		if err != nil {
			return err, nil
		}
	}

	safeRef = SafeSproutsReference{
		BasePath: sprouts.BasePath,
		Garden: sprouts.Garden,
	}

	return nil, &safeRef 
}