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
			var rootPath string
			var err error

			if len(args) > 0 {
				rootPath = args[0]
			}

			var relative bool = true

			if options["relative"] != nil {
				relative = options["relative"].(bool)
			} else {
				relative = true
			}

			err, rootPath = dydfs.PartialEvalSymlinks(task.SERIAL_CONTEXT, rootPath)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving root path")
				return 1
			}

			err, garden := dryad.Garden(rootPath).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden")
				return 1
			}

			err, roots := garden.Roots().Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden roots")
				return 1
			}
	
			err, safeRoot := roots.Root(rootPath).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving root")
				return 1
			}

			err, safeRequirements := safeRoot.Requirements().Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving root requirements")
				return 1
			} else if safeRequirements == nil {
				// no requirements, so exit
				return 0				
			}

			var onRequirementMatch =func (ctx *task.ExecutionContext, requirement *dryad.SafeRootRequirementReference) (error, any) {
				zlog.Trace().
					Str("path", requirement.BasePath).
					Msg("root requirements list / onRequirement")				

				if relative {
					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(
						requirement.Requirements.Root.Roots.Garden.BasePath,
						requirement.BasePath,
					)
					if err != nil {
						return err, nil
					}
					fmt.Println(relPath)
				} else {
					fmt.Println(requirement.BasePath)
				}
				return nil, nil
			}

			err = safeRequirements.Walk(
				task.SERIAL_CONTEXT,
				dryad.RootRequirementsWalkRequest{
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
