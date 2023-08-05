package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	log "github.com/rs/zerolog/log"
)

var scriptsListAction = func(req clib.ActionRequest) int {
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

	var showPath bool

	if options["path"] != nil {
		showPath = options["path"].(bool)
	} else {
		showPath = false
	}

	// if the scope is unset, bypass expansion and run the action directly
	if scope == "" || scope == "none" {
		log.Fatal().Msg("no scope set, can't find command")
		return 1
	} else {
		log.Info().Msg("using scope: " + scope)
	}

	var scripts []string

	err = dryad.ScriptsWalk(dryad.ScriptsWalkRequest{
		BasePath: basePath,
		Scope:    scope,
		OnMatch: func(path string, info fs.FileInfo) error {
			if showPath {
				scripts = append(scripts, path)
			} else {
				var name string = info.Name()
				var script string = "dryad script run " + strings.TrimPrefix(name, "script-run-")
				scripts = append(scripts, script)
			}
			return nil
		},
	})
	if err != nil {
		log.Fatal().Err(err)
		return 1
	}

	sort.Strings(scripts)

	for _, script := range scripts {
		fmt.Println(script)
	}

	return 0
}

var scriptsListCommand = func() clib.Command {
	command := clib.NewCommand("list", "list all available scripts in the current scope").
		WithOption(clib.NewOption("scope", "set the scope for the command")).
		WithOption(clib.NewOption("path", "print the path to the scripts instead of the script run command").WithType(clib.OptionTypeBool)).
		WithAction(scopeHandler(scriptsListAction))

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()
