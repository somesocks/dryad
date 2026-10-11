package core

import (

	// "errors"

	"context"
	"io/ioutil"
	// "dryad/internal/filepath"
	// "os"
	// zlog "github.com/rs/zerolog/log"
)

func (sproutTrait *SafeSproutTraitReference) Get(ctx context.Context) (error, string) {
	bytes, err := ioutil.ReadFile(sproutTrait.BasePath)
	if err != nil {
		return err, ""
	}

	return nil, string(bytes)
}
