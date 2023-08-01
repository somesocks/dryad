package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
)

var scriptPathAction = func(req clib.ActionRequest) int {
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

	scriptPath, err := dryad.ScriptPath(dryad.ScriptPathRequest{
		BasePath: basePath,
		Scope:    scope,
		Setting:  "script-run-" + command,
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(scriptPath)

	return 0
}

var scriptPathCommand = clib.NewCommand("path", "print the path to a script").
	WithArg(
		clib.
			NewArg("command", "the script name").
			WithType(clib.ArgTypeString).
			WithAutoComplete(ArgAutoCompleteScript),
	).
	WithOption(clib.NewOption("scope", "set the scope for the command")).
	WithAction(scopeHandler(scriptPathAction))
