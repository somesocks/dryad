package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"os"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

var scriptRunAction = func(req clib.ActionRequest) int {
	var command = req.Args[0]
	var args = req.Args[1:]
	var options = req.Opts

	basePath, err := os.Getwd()
	if err != nil {
		zlog.Fatal().Err(err).Msg("error finding working directory")
		return 1
	}

	gardenPath, err := dryad.GardenPath(basePath)
	if err != nil {
		zlog.Fatal().Err(err).Msg("error finding garden path")
		return 1
	}

	var scope string
	if options["scope"] != nil {
		scope = options["scope"].(string)
	} else {
		var err error
		scope, err = dryad.ScopeGetDefault(scope)
		zlog.Info().Msg("loading default scope: " + scope)
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
		zlog.Info().Msg("using scope: " + scope)
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
		GardenPath: gardenPath,
		Scope:      scope,
		Setting:    "script-run-" + command,
		Args:       args,
		Env:        env,
	})
	if err != nil {
		zlog.Fatal().Err(err).Msg("error while running script")
		return 1
	}

	return 0
}

var scriptRunCommand = func() clib.Command {
	command := clib.NewCommand("run", "run a script in the current scope").
		WithArg(
			clib.
				NewArg("command", "the script name").
				WithType(clib.ArgTypeString).
				WithAutoComplete(ArgAutoCompleteScript),
		).
		WithOption(clib.NewOption("inherit", "pass all environment variables from the parent environment to the alias to exec").WithType(clib.OptionTypeBool)).
		WithArg(clib.NewArg("-- args", "args to pass to the script").AsOptional()).
		WithAction(scriptRunAction)

	command = ScopedCommand(command)
	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()
