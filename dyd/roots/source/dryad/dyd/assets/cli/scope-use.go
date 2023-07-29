package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
	"os"
)

var scopeUseCommand = clib.NewCommand("use", "set a scope to be active. alias for `dryad scopes default set`").
	WithArg(clib.NewArg("name", "the name of the scope to set as active. use 'none' to unset the active scope")).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args

		var name string = args[0]

		var path, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		if name == "none" {
			err = dryad.ScopeUnsetDefault(path)
		} else {
			err = dryad.ScopeSetDefault(path, name)
		}

		if err != nil {
			log.Fatal(err)
		}

		return 0
	})
