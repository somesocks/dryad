package core

import (

	// "errors"

	"context"
	"io/ioutil"
	// "dryad/internal/filepath"
	// "os"
	// zlog "github.com/rs/zerolog/log"
)

func (rootTrait *SafeRootTraitReference) Get(ctx context.Context) (error, string) {
	bytes, err := ioutil.ReadFile(rootTrait.BasePath)
	if err != nil {
		return err, ""
	}

	return nil, string(bytes)
}
