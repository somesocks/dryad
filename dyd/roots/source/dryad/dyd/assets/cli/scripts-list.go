package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

var scriptsListAction = func(req clib.ActionRequest) int {
	var options = req.Opts

	basePath, err := os.Getwd()
	if err != nil {
		zlog.Fatal().Err(err).Msg("error while finding working directory")
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

	var showPath bool = false
	if options["path"] != nil {
		showPath = options["path"].(bool)
	}

	var showOneline bool = true
	if options["oneline"] != nil {
		showOneline = options["oneline"].(bool)
	}

	// if the scope is unset, bypass expansion and run the action directly
	if scope == "" || scope == "none" {
		zlog.Fatal().Msg("no scope set, can't find command")
		return 1
	} else {
		zlog.Info().Msg("using scope: " + scope)
	}

	var scripts []string

	err = dryad.ScriptsWalk(dryad.ScriptsWalkRequest{
		BasePath: basePath,
		Scope:    scope,
		OnMatch: func(path string, info fs.FileInfo) error {
			if showPath {
				scripts = append(scripts, path)
			} else {
				name := info.Name()
				script := "dryad run " + strings.TrimPrefix(name, "script-run-")

				if showOneline {
					oneline, _ := dryad.ScriptOnelineGet(path)
					if oneline != "" {
						script = script + " - " + oneline
					}
				}

				scripts = append(scripts, script)
			}
			return nil
		},
	})
	if err != nil {
		zlog.Fatal().Err(err).Msg("error while crawling scripts")
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
		WithOption(clib.NewOption("path", "print the path to the scripts instead of the script run command").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("oneline", "print the oneline decription of each command").WithType(clib.OptionTypeBool)).
		WithAction(scriptsListAction)

	command = ScopedCommand(command)
	command = LoggingCommand(command)


	return command
}()
