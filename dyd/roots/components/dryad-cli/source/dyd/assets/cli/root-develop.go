package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	dydfs "dryad/filesystem"
	"fmt"

	zlog "github.com/rs/zerolog/log"
)

var rootDevelopCommand = func() clib.Command {
	command := clib.NewCommand("develop", "create a temporary development environment for a root").
		WithArg(
			clib.
				NewArg("path", "path to the root to develop").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("editor", "choose the editor to run in the root development environment").WithType(clib.OptionTypeString)).
		WithOption(clib.NewOption("arg", "argument to pass to the editor").WithType(clib.OptionTypeMultiString)).
		WithOption(clib.NewOption("inherit", "inherit env variables from the host environment").WithType(clib.OptionTypeBool)).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args
			var opts = req.Opts

			var path string
			var editor string
			var editorArgs []string
			var inherit bool
			var err error

			if len(args) > 0 {
				path = args[0]
			}

			if opts["editor"] != nil {
				editor = opts["editor"].(string)
			} else {
				editor = ""
			}

			if opts["arg"] != nil {
				editorArgs = opts["arg"].([]string)
			}

			if opts["inherit"] != nil {
				inherit = opts["inherit"].(bool)
			}

			err, path = dydfs.PartialEvalSymlinks(task.SERIAL_CONTEXT, path)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving root path")
				return 1
			}

			err, garden := dryad.Garden(path).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden")
				return 1
			}	

			err, roots := garden.Roots().Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden roots")
				return 1
			}

			err, safeRootRef := roots.Root(path).Resolve(task.SERIAL_CONTEXT, nil)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving root")
				return 1
			}	

			var rootFingerprint string
			err, rootFingerprint = safeRootRef.Develop(
				task.SERIAL_CONTEXT,
				dryad.RootDevelopRequest{
					Editor: editor,
					EditorArgs: editorArgs,
					Inherit: inherit,
				},
			)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error from root development environment")
				return 1
			}
			fmt.Println(rootFingerprint)

			return 0
		})

	command = ScopedCommand(command)
	command = LoggingCommand(command)


	return command
}()
