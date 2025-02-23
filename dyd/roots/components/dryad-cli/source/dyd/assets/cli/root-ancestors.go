package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	dydfs "dryad/filesystem"
	"fmt"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var rootAncestorsCommand = func() clib.Command {

	type ParsedArgs struct {
		RootPath string
		Relative bool
		Parallel int
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = 
		func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts
			var err error 

			var rootPath string

			if len(args) > 0 {
				rootPath = args[0]
			}

			var relative bool = true

			if options["relative"] != nil {
				relative = options["relative"].(bool)
			} else {
				relative = true
			}

			var parallel int

			if options["parallel"] != nil {
				parallel = int(options["parallel"].(int64))
			} else {
				parallel = 8
			}

			err, rootPath = dydfs.PartialEvalSymlinks(ctx, rootPath)
			if err != nil {
				return err, ParsedArgs{}
			}
	
			return nil, ParsedArgs{
				RootPath: rootPath,
				Relative: relative,
				Parallel: parallel,
			}
		}

	var findAncestors = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		var rootPath string = args.RootPath
		var relative bool = args.Relative
		
		err, garden := dryad.Garden(args.RootPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, root := roots.Root(rootPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}
		rootPath = root.BasePath

		err, graph := roots.Graph(
			ctx,
			dryad.RootsGraphRequest{
				Relative: relative,
			},
		)
		if err != nil {
			return err, nil
		}

		if relative {
			rootPath, err = filepath.Rel(garden.BasePath, rootPath)
			if err != nil {
				return err, nil
			}
		}

		ancestors := graph.Descendants(make(dryad.TStringSet), []string{rootPath}).ToArray([]string{})

		for _, v := range ancestors {
			fmt.Println(v)
		}

		return nil, nil
	}

	findAncestors = task.WithContext(
		findAncestors,
		func (ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			findAncestors,
		),
		func (err error, val any) int {
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding root ancestors")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("ancestors", "list all roots the selected root depends on (directly and indirectly)").
		WithArg(
			clib.
				NewArg("root_path", "path to the root").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("relative", "print roots relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithAction(action)

	command = ParallelCommand(command)
	command = LoggingCommand(command)

	return command
}()
