package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
)

var scopeSettingGetCommand = clib.NewCommand("get", "print the value of a setting in a scope, if it exists").
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

		value, err := dryad.ScopeSettingGet(path, scope, setting)
		if err != nil {
			log.Fatal(err)
		}

		if value != "" {
			fmt.Println(value)
		}

		return 0
	})
