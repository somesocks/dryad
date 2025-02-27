
package core

import (
	"dryad/task"
	// "errors"

	"io/ioutil"

	// "path/filepath"

	// "os"

	// zlog "github.com/rs/zerolog/log"
)


func (rootTrait *SafeRootTraitReference) Get(ctx * task.ExecutionContext) (error, string) {
	bytes, err := ioutil.ReadFile(rootTrait.BasePath)
	if err != nil {
		return err, ""
	}

	return nil, string(bytes)
}