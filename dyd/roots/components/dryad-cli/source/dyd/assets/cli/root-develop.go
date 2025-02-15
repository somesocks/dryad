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

			if !filepath.IsAbs(path) {
				wd, err := os.Getwd()
				if err != nil {
					zlog.Fatal().Err(err).Msg("error while finding working directory")
					return 1
				}
				path = filepath.Join(wd, path)
			}

			unsafeGarden := dryad.UnsafeGardenReference{
				BasePath: path,
			}
			
			err, garden := unsafeGarden.Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden")
				return 1
			}	

			unsafeRootRef := dryad.UnsafeRootReference{
				Garden: garden,
				BasePath: path,
			}

			err, safeRootRef := unsafeRootRef.Resolve(task.SERIAL_CONTEXT, nil)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving root")
				return 1
			}	

			var rootFingerprint string
			rootFingerprint, err = dryad.RootDevelop(
				task.SERIAL_CONTEXT,
				dryad.RootDevelopRequest{
					Root: &safeRootRef,
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
