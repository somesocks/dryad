
package core

import (
	// fs2 "dryad/filesystem"
	"dryad/task"

	// "os"

	"errors"
	// zlog "github.com/rs/zerolog/log"
)

var ErrorNoSproutTraits = errors.New("sprout does not have traits")

func (sproutTraits *UnsafeSproutTraitsReference) Resolve(ctx * task.ExecutionContext) (error, *SafeSproutTraitsReference) {
	var sproutTraitsExists bool
	var err error
	var safeRef SafeSproutTraitsReference

	sproutTraitsExists, err = fileExists(sproutTraits.BasePath)
	if err != nil {
		return err, nil
	}

	if !sproutTraitsExists {
		return ErrorNoSproutTraits, nil
	}

	safeRef = SafeSproutTraitsReference{
		BasePath: sproutTraits.BasePath,
		Sprout: sproutTraits.Sprout,
	}

	return nil, &safeRef 
}