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

var rootDescendantsCommand = func() clib.Command {
	command := clib.NewCommand("descendants", "list all roots that depend on the selected root (directly and indirectly)").
		WithArg(
			clib.
				NewArg("root_path", "path to the root").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("relative", "print roots relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithAction(func(req clib.ActionRequest) int {
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
					zlog.Fatal().Err(err).Msg("error while finding working directory")
					return 1
				}
				rootPath = filepath.Join(wd, rootPath)
			}


			unsafeGarden := dryad.Garden(rootPath)
			
			err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden")
				return 1
			}
	
			rootPath, err = dryad.RootPath(rootPath, "")
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving root path")
				return 1
			}

			graph, err := dryad.RootsGraph(
				dryad.RootsGraphRequest{
					Garden: garden,
					Relative: relative,
				},
			)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while building graph")
				return 1
			}

			graph = graph.Transpose()

			if relative {
				rootPath, err = filepath.Rel(garden.BasePath, rootPath)
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while making root relative")
					return 1
				}
			}

			ancestors := graph.Descendants(make(dryad.TStringSet), []string{rootPath}).ToArray([]string{})
			for _, v := range ancestors {
				fmt.Println(v)
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()
