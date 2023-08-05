package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"os"

	log "github.com/rs/zerolog/log"
)

var scriptPathAction = func(req clib.ActionRequest) int {
	var command = req.Args[0]
	var options = req.Opts

	basePath, err := os.Getwd()
	if err != nil {
		log.Fatal().Err(err)
		return 1
	}

	var scope string
	if options["scope"] != nil {
		scope = options["scope"].(string)
	} else {
		var err error
		scope, err = dryad.ScopeGetDefault(scope)
		log.Info().Msg("loading default scope: " + scope)
		if err != nil {
			log.Fatal().Err(err)
			return 1
		}
	}

	// if the scope is unset, bypass expansion and run the action directly
	if scope == "" || scope == "none" {
		log.Fatal().Msg("no scope set, can't find command")
		return 1
	} else {
		log.Info().Msg("using scope: " + scope)
	}

	scriptPath, err := dryad.ScriptPath(dryad.ScriptPathRequest{
		BasePath: basePath,
		Scope:    scope,
		Setting:  "script-run-" + command,
	})
	if err != nil {
		log.Fatal().Err(err)
		return 1
	}

	fmt.Println(scriptPath)

	return 0
}

var scriptPathCommand = func() clib.Command {
	command := clib.NewCommand("path", "print the path to a script").
		WithArg(
			clib.
				NewArg("command", "the script name").
				WithType(clib.ArgTypeString).
				WithAutoComplete(ArgAutoCompleteScript),
		).
		WithOption(clib.NewOption("scope", "set the scope for the command")).
		WithAction(scopeHandler(scriptPathAction))

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()
