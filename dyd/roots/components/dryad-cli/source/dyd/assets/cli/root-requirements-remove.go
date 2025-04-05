package cli

import (
	clib "dryad/cli-builder"
	core "dryad/core"
	task "dryad/task"
	
	"os"

	zlog "github.com/rs/zerolog/log"
)

var rootRequirementsRemoveCommand = func() clib.Command {
	command := clib.NewCommand("remove", "remove a requirement from the current root").
		WithArg(
			clib.
				NewArg("name", "name of the requirement to remove").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var requirementName = args[0]

			var rootPath, err = os.Getwd()
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while finding working directory")
				return 1
			}

			err, garden := core.Garden(rootPath).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving garden")
				return 1
			}

			err, roots := garden.Roots().Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving roots")
				return 1
			}

			err, root := roots.Root(rootPath).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving root")
				return 1
			}
			
			err, requirements := root.
				Requirements().
				Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving root requirements")
				return 1
			} else if requirements == nil {
				zlog.Fatal().Err(err).Msg("root has no requirements")
				return 1
			}

			err, requirement := requirements.
				Requirement(requirementName).
				Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving requirement")
				return 1
			} else if requirement == nil {
				zlog.Fatal().Err(err).Msg("requirement does not exist")
				return 1
			}

			err = requirement.Remove(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while unlinking root")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()
