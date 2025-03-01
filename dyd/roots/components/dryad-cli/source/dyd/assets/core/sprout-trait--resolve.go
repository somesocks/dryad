
package core

import (
	"dryad/task"
	"errors"

	// "path/filepath"

	// "os"

	// zlog "github.com/rs/zerolog/log"
)

var ErrorNoSproutTrait = errors.New("sprout does not have trait")

func (sproutTrait *UnsafeSproutTraitReference) Resolve(ctx * task.ExecutionContext) (error, *SafeSproutTraitReference) {
	var sproutTraitExists bool
	var err error
	var safeRef SafeSproutTraitReference

	sproutTraitExists, err = fileExists(sproutTrait.BasePath)
	if err != nil {
		return err, nil
	}

	if !sproutTraitExists {
		return ErrorNoSproutTrait, nil
	}

	safeRef = SafeSproutTraitReference{
		BasePath: sproutTrait.BasePath,
		Traits: sproutTrait.Traits,
	}

	return nil, &safeRef 
}