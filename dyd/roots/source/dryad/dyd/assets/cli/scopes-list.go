package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"io/fs"
	"log"
	"path/filepath"
	"sort"
)

var scopesListCommand = clib.NewCommand("list", "list all scopes in the current garden").
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args

		var path string = ""
		var err error

		if len(args) > 0 {
			path = args[0]
		}

		var scopes []string

		err = dryad.ScopesWalk(path, func(path string, info fs.FileInfo) error {
			scopes = append(scopes, filepath.Base(path))
			return nil
		})
		if err != nil {
			log.Fatal(err)
		}

		sort.Strings(scopes)

		for _, scope := range scopes {
			fmt.Println(scope)
		}

		return 0
	})
