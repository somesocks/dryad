package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
)

var scriptGetAction = func(req clib.ActionRequest) int {
	var command = req.Args[0]
	var options = req.Opts

	basePath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var scope string
	if options["scope"] != nil {
		scope = options["scope"].(string)
	} else {
		var err error
		scope, err = dryad.ScopeGetDefault(scope)
		fmt.Println("[info] loading default scope:", scope)
		if err != nil {
			log.Fatal(err)
		}
	}

	// if the scope is unset, bypass expansion and run the action directly
	if scope == "" || scope == "none" {
		log.Fatal("no scope set, can't find command")
	} else {
		fmt.Println("[info] using scope:", scope)
	}

	script, err := dryad.ScriptGet(dryad.ScriptGetRequest{
		BasePath: basePath,
		Scope:    scope,
		Setting:  "script-run-" + command,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(script)

	return 0
}

var scriptGetCommand = clib.NewCommand("get", "print the contents of a script").
	WithArg(clib.NewArg("command", "the script name").WithType(clib.ArgTypeString)).
	WithOption(clib.NewOption("scope", "set the scope for the command")).
	WithAction(scopeHandler(scriptGetAction))
