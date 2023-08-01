package cli

import (
	dryad "dryad/core"
	"io/fs"
	"log"
	"os"
	"strings"
)

func ArgAutoCompleteScript(token string) []string {
	var results = []string{}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	activeScope, err := dryad.ScopeGetDefault(wd)
	if err != nil {
		log.Fatal(err)
	}

	if activeScope == "" {
		return results
	}

	err = dryad.ScriptsWalk(dryad.ScriptsWalkRequest{
		BasePath: wd,
		Scope:    activeScope,
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
		log.Fatal(err)
	}

	return results
}