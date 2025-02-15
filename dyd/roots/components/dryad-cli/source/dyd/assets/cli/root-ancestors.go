package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"os"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var rootAncestorsCommand = func() clib.Command {

	type ParsedArgs struct {
		RootPath string
		Relative bool
	}

	var parseArgs = task.From(
		func(req clib.ActionRequest) (error, ParsedArgs) {
			var args = req.Args
			var options = req.Opts

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

			if !filepath.IsAbs(rootPath) {
				wd, err := os.Getwd()
				if err != nil {
					return err, ParsedArgs{}
				}
				rootPath = filepath.Join(wd, rootPath)
			}
	
			return nil, ParsedArgs{
				RootPath: rootPath,
				Relative: relative,
			}
		},
	)

	var findAncestors = func (ctx *task.ExecutionContext, args ParsedArgs) (error, any) {

		rootPath := args.RootPath
		relative := args.Relative

		unsafeGarden := dryad.Garden(args.RootPath)
		
		err, garden := unsafeGarden.Resolve(ctx)
		if err != nil {
			return err, nil
		}

		rootPath, err = dryad.RootPath(rootPath, "")
		if err != nil {
			return err, nil
		}

		graph, err := dryad.RootsGraph(
			dryad.RootsGraphRequest{
				Garden: garden,
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
			return nil, ctx
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

	command = LoggingCommand(command)


	return command
}()
