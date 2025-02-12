package cli

import (
	dryad "dryad/core"
	"dryad/task"
	"io/fs"
	"path/filepath"
	"strings"
)

func ArgAutoCompleteScope(token string) (error, []string) {
	var results = []string{}

	unsafeGarden := dryad.UnsafeGardenReference{
		BasePath: "",
	}
	
	err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT, nil)
	if err != nil {
		return err, results
	}

	err = dryad.ScopesWalk(&garden, func(path string, info fs.FileInfo) error {
		var scope = filepath.Base(path)

		if strings.HasPrefix(scope, token) {
			results = append(results, scope)
		}

		return nil
	})
	if err != nil {
		return err, results
	}

	return nil, results
}
