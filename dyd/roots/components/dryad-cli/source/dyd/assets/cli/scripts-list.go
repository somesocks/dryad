package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"sort"
	"strings"

	zlog "github.com/rs/zerolog/log"
)

var scriptsListAction = func(req clib.ActionRequest) int {
	type ParsedArgs struct {
		Scope       string
		HasScope    bool
		ShowPath    bool
		ShowOneline bool
		GardenPath  string
		Parallel    int
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var opts = req.Opts
			var parallel int
			var scope string
			var hasScope bool
			var showPath bool
			var showOneline bool = true

			if opts["parallel"] != nil {
				parallel = int(opts["parallel"].(int64))
			} else {
				parallel = PARALLEL_COUNT_DEFAULT
			}

			if opts["scope"] != nil {
				scope = opts["scope"].(string)
				hasScope = true
			}

			if opts["path"] != nil {
				showPath = opts["path"].(bool)
			}

			if opts["oneline"] != nil {
				showOneline = opts["oneline"].(bool)
			}

			path, err := os.Getwd()
			if err != nil {
				return err, ParsedArgs{}
			}

			return nil, ParsedArgs{
				Scope:       scope,
				HasScope:    hasScope,
				ShowPath:    showPath,
				ShowOneline: showOneline,
				GardenPath:  path,
				Parallel:    parallel,
			}
		},
	)

	var listScripts = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		unsafeGarden := dryad.Garden(args.GardenPath)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		scope := args.Scope
		if !args.HasScope {
			scope, err = dryad.ScopeGetDefault(garden)
			zlog.Debug().Msg("loading default scope: " + scope)
			if err != nil {
				return err, nil
			}
		}

		if scope == "" || scope == "none" {
			return errors.New("no scope set, can't find command"), nil
		}
		zlog.Debug().Msg("using scope: " + scope)

		var scripts []string

		err = dryad.ScriptsWalk(dryad.ScriptsWalkRequest{
			Garden: garden,
			Scope:  scope,
			OnMatch: func(path string, info fs.FileInfo) error {
				if args.ShowPath {
					scripts = append(scripts, path)
				} else {
					name := info.Name()
					script := "dryad run " + strings.TrimPrefix(name, "script-run-")

					if args.ShowOneline {
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
			return err, nil
		}

		sort.Strings(scripts)

		for _, script := range scripts {
			fmt.Println(script)
		}

		return nil, nil
	}

	listScripts = task.WithContext(
		listScripts,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	return task.Return(
		task.Series2(
			parseArgs,
			listScripts,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while crawling scripts")
				return 1
			}
			return 0
		},
	)(req)
}

var scriptsListCommand = func() clib.Command {
	command := clib.NewCommand("list", "list all available scripts in the current scope").
		WithOption(clib.NewOption("path", "print the path to the scripts instead of the script run command").WithType(clib.OptionTypeBool)).
		WithOption(clib.NewOption("oneline", "print the oneline decription of each command").WithType(clib.OptionTypeBool)).
		WithAction(scriptsListAction)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
