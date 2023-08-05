package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"os"
	"strings"

	log "github.com/rs/zerolog/log"
)

var scriptEditAction = func(req clib.ActionRequest) int {
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
		log.Fatal().Err(err)
		return 1
	}

	return 0
}

var scriptEditCommand = func() clib.Command {
	command := clib.NewCommand("edit", "edit a script").
		WithArg(
			clib.
				NewArg("command", "the script name").
				WithType(clib.ArgTypeString).
				WithAutoComplete(ArgAutoCompleteScript),
		).
		WithOption(clib.NewOption("editor", "set the editor to use")).
		WithAction(scriptEditAction)

	command = ScopedCommand(command)
	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()
