package cli

import (
	dryad "dryad/core"
	"dryad/task"
	"io/fs"
	"os"
	"strings"
)

func ArgAutoCompleteScript(token string) (error, []string) {
	var results = []string{}

	wd, err := os.Getwd()
	if err != nil {
		return err, results
	}

	unsafeGarden := dryad.UnsafeGardenReference{
		BasePath: wd,
	}
	
	err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT)
	if err != nil {
		return err, results
	}

	activeScope, err := dryad.ScopeGetDefault(garden)
	if err != nil {
		return err, results
	}

	if activeScope == "" {
		return nil, results
	}

	err = dryad.ScriptsWalk(dryad.ScriptsWalkRequest{
		Garden: garden,
		Scope: activeScope,
		OnMatch: func(path string, info fs.FileInfo) error {
			var name string = info.Name()
			var script string = strings.TrimPrefix(name, "script-run-")
			if strings.HasPrefix(script, token) {
				results = append(results, script)
			}
			return nil
		},
	})
	if err != nil {
		return err, results
	}

	return nil, results
}
