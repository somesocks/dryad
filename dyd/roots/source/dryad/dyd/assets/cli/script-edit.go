package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
	"strings"
)

var scriptEditAction = func(req clib.ActionRequest) int {
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

	var env = map[string]string{}

	for _, e := range os.Environ() {
		if i := strings.Index(e, "="); i >= 0 {
			env[e[:i]] = e[i+1:]
		}
	}

	if options["editor"] != nil {
		editor := options["editor"].(string)
		env["EDITOR"] = editor
	}

	err = dryad.ScriptEdit(dryad.ScriptEditRequest{
		BasePath: basePath,
		Scope:    scope,
		Setting:  "script-run-" + command,
		Env:      env,
	})
	if err != nil {
		log.Fatal(err)
	}

	return 0
}

var scriptEditCommand = clib.NewCommand("edit", "edit a script").
	WithArg(
		clib.
			NewArg("command", "the script name").
			WithType(clib.ArgTypeString).
			WithAutoComplete(ArgAutoCompleteScript),
	).
	WithOption(clib.NewOption("scope", "set the scope for the command")).
	WithOption(clib.NewOption("editor", "set the editor to use")).
	WithAction(scopeHandler(scriptEditAction))
