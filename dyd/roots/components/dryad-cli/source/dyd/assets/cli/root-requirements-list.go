package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	"fmt"
	"path/filepath"

	zlog "github.com/rs/zerolog/log"
)

var rootRequirementsListCommand = func() clib.Command {
	command := clib.NewCommand("list", "list all requirements of this root").
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
			var root string

			if len(args) > 0 {
				root = args[0]
			}

			var relative bool = true

			if options["relative"] != nil {
				relative = options["relative"].(bool)
			} else {
				relative = true
			}

			unsafeGarden := dryad.UnsafeGardenReference{
				BasePath: root,
			}
			
			err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden")
				return 1
			}

			unsafeRoot := dryad.UnsafeRootReference{
				BasePath: root,
				Garden: garden,
			}
	
			err, safeRoot := unsafeRoot.Resolve(task.SERIAL_CONTEXT, nil)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving root")
				return 1
			}

			var onRequirementMatch =func (ctx *task.ExecutionContext, match *dryad.SafeRootReference) (error, any) {
				zlog.Trace().
					Str("path", match.BasePath).
					Msg("root requirements list / onRequirement")				

				if relative {
					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(match.Garden.BasePath, match.BasePath)
					if err != nil {
						return err, nil
					}
					fmt.Println(relPath)
				} else {
					fmt.Println(match.BasePath)
				}
				return nil, nil
			}

			err, _ = dryad.RootRequirementsWalk(
				task.SERIAL_CONTEXT,
				dryad.RootRequirementsWalkRequest{
					Root: &safeRoot,
					OnMatch: onRequirementMatch,
				},
			)

			if err != nil {
				zlog.Fatal().Err(err).Msg("error while crawling root requirements")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()
