package cli

import (
	dryad "dryad/core"
	"io/fs"
	"log"
	"path/filepath"
	"strings"
)

func ArgAutoCompleteScope(token string) []string {
	var results = []string{}

	err := dryad.ScopesWalk("", func(path string, info fs.FileInfo) error {
		var scope = filepath.Base(path)

		if strings.HasPrefix(scope, token) {
			results = append(results, scope)
		}

		return nil
	})
	if err != nil {
		log.Fatal(err)
	}

	return results
}
