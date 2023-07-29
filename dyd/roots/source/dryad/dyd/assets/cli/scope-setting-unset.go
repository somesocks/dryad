package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
	"os"
)

var scopeSettingUnsetCommand = clib.NewCommand("unset", "remove a setting from a scope").
	WithArg(clib.NewArg("scope", "the name of the scope")).
	WithArg(clib.NewArg("setting", "the name of the setting")).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args

		var scope string = args[0]
		var setting string = args[1]

		var path, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		err = dryad.ScopeSettingUnset(path, scope, setting)
		if err != nil {
			log.Fatal(err)
		}

		return 0
	})
