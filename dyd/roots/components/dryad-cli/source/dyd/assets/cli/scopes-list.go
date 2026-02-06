package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"

	zlog "github.com/rs/zerolog/log"
)

var scopesListCommand = func() clib.Command {
	type ParsedArgs struct {
		Path     string
		Oneline  bool
		Parallel int
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var opts = req.Opts

			var path string
			var oneline bool = true
			var parallel int

			if len(args) > 0 {
				path = args[0]
			} else {
				cwd, err := os.Getwd()
				if err != nil {
					return err, ParsedArgs{}
				}
				path = cwd
			}

			if opts["oneline"] != nil {
				oneline = opts["oneline"].(bool)
			}

			if opts["parallel"] != nil {
				parallel = int(opts["parallel"].(int64))
			} else {
				parallel = PARALLEL_COUNT_DEFAULT
			}

			return nil, ParsedArgs{
				Path:     path,
				Oneline:  oneline,
				Parallel: parallel,
			}
		},
	)

	var listScopes = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		var scopes []string

		unsafeGarden := dryad.Garden(args.Path)

		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err = dryad.ScopesWalk(garden, func(path string, info fs.FileInfo) error {
			scope := filepath.Base(path)

			// fetch the oneline if enabled
			if args.Oneline {
				scopeOneline, _ := dryad.ScopeOnelineGet(garden, scope)
				if scopeOneline != "" {
					scope = scope + " - " + scopeOneline
				}
			}

			scopes = append(scopes, scope)

			return nil
		})
		if err != nil {
			return err, nil
		}

		sort.Strings(scopes)

		for _, scope := range scopes {
			fmt.Println(scope)
		}

		return nil, nil
	}

	listScopes = task.WithContext(
		listScopes,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			listScopes,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while crawling scopes")
				return 1
			}
			return 0
		},
	)

	command := clib.NewCommand("list", "list all scopes in the current garden").
		WithOption(clib.NewOption("oneline", "enable/disable printing one-line scope descriptions").WithType(clib.OptionTypeBool)).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()
