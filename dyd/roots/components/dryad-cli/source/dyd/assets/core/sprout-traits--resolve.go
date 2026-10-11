package core

import (

	// "os"

	"context"
	"errors"
	// zlog "github.com/rs/zerolog/log"
)

var ErrorNoSproutTraits = errors.New("sprout does not have traits")

func (sproutTraits *UnsafeSproutTraitsReference) Resolve(ctx context.Context) (error, *SafeSproutTraitsReference) {
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
		Sprout:   sproutTraits.Sprout,
	}

	return nil, &safeRef
}
