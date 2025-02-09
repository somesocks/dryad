package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"

	zlog "github.com/rs/zerolog/log"
)

var scriptGetAction = func(req clib.ActionRequest) int {
	var command = req.Args[0]
	var options = req.Opts

	basePath, err := os.Getwd()
	if err != nil {
		zlog.Fatal().Err(err).Msg("error while finding working directory")
		return 1
	}

	unsafeGarden := dryad.UnsafeGardenReference{
		BasePath: basePath,
	}
	
	err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT, nil)
	if err != nil {
		return 1
	}

	var scope string
	if options["scope"] != nil {
		scope = options["scope"].(string)
	} else {
		var err error
		scope, err = dryad.ScopeGetDefault(&garden)
		zlog.Debug().Msg("loading default scope: " + scope)
		if err != nil {
			zlog.Fatal().Err(err).Msg("error while finding active scope")
			return 1
		}
	}

	// if the scope is unset, bypass expansion and run the action directly
	if scope == "" || scope == "none" {
		zlog.Fatal().Msg("no scope set, can't find command")
		return 1
	} else {
		zlog.Debug().Msg("using scope: " + scope)
	}

	script, err := dryad.ScriptGet(dryad.ScriptGetRequest{
		Garden: &garden,
		Scope:    scope,
		Setting:  "script-run-" + command,
	})
	if err != nil {
		zlog.Fatal().Err(err).Msg("error while finding script")
		return 1
	}

	fmt.Println(script)

	return 0
}

var scriptGetCommand = func() clib.Command {
	command := clib.NewCommand("get", "print the contents of a script").
		WithArg(
			clib.
				NewArg("command", "the script name").
				WithType(clib.ArgTypeString).
				WithAutoComplete(ArgAutoCompleteScript),
		).
		WithAction(scriptGetAction)

	command = ScopedCommand(command)
	command = LoggingCommand(command)


	return command
}()
