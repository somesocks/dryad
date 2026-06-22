package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	// "os"

	zlog "github.com/rs/zerolog/log"
)

func resolveSproutsReference(ctx *task.ExecutionContext, sprouts *UnsafeSproutsReference) (error, *SafeSproutsReference) {
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
		Garden:   sprouts.Garden,
	}

	return nil, &safeRef
}

var memoResolveSproutsReference = task.Memoize(
	resolveSproutsReference,
	func(ctx *task.ExecutionContext, sprouts *UnsafeSproutsReference) (error, any) {
		type Key struct {
			Group      string
			BasePath   string
			GardenPath string
		}

		gardenPath := ""
		if sprouts.Garden != nil {
			gardenPath = sprouts.Garden.BasePath
		}

		return nil, Key{
			Group:      "Sprouts.Resolve",
			BasePath:   sprouts.BasePath,
			GardenPath: gardenPath,
		}
	},
)

func (sprouts *UnsafeSproutsReference) Resolve(ctx *task.ExecutionContext) (error, *SafeSproutsReference) {
	return memoResolveSproutsReference(ctx, sprouts)
}
