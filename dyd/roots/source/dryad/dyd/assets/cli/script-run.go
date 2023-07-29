package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
	"strings"
)

var scriptRunAction = func(req clib.ActionRequest) int {
	var command = req.Args[0]
	var args = req.Args[1:]
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

	var inherit bool
	var env = map[string]string{}

	if options["inherit"] != nil {
		inherit = options["inherit"].(bool)
	} else {
		inherit = true
	}

	// pull environment variables from parent process
	if inherit {
		for _, e := range os.Environ() {
			if i := strings.Index(e, "="); i >= 0 {
				env[e[:i]] = e[i+1:]
			}
		}
	} else {
		// copy a few variables over from parent env for convenience
		env["TERM"] = os.Getenv("TERM")
	}

	err = dryad.ScriptRun(dryad.ScriptRunRequest{
		BasePath: basePath,
		Scope:    scope,
		Setting:  "script-run-" + command,
		Args:     args,
		Env:      env,
	})
	if err != nil {
		log.Fatal(err)
	}

	return 0
}

var scriptRunCommand = clib.NewCommand("run", "run a script in the current scope").
	WithArg(clib.NewArg("command", "the script name").WithType(clib.ArgTypeString)).
	WithOption(clib.NewOption("scope", "set the scope for the command")).
	WithOption(clib.NewOption("inherit (default true)", "pass all environment variables from the parent environment to the alias to exec").WithType(clib.OptionTypeBool)).
	WithArg(clib.NewArg("-- args", "args to pass to the script").AsOptional()).
	WithAction(scopeHandler(scriptRunAction))
