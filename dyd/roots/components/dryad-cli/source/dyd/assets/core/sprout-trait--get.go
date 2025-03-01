
package core

import (
	"dryad/task"
	// "errors"

	"io/ioutil"

	// "path/filepath"

	// "os"

	// zlog "github.com/rs/zerolog/log"
)


func (sproutTrait *SafeSproutTraitReference) Get(ctx * task.ExecutionContext) (error, string) {
	bytes, err := ioutil.ReadFile(sproutTrait.BasePath)
	if err != nil {
		return err, ""
	}

	return nil, string(bytes)
}