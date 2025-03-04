package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"dryad/task"
	dydfs "dryad/filesystem"

	zlog "github.com/rs/zerolog/log"
)

var rootReplaceCommand = func() clib.Command {
	command := clib.NewCommand("replace", "replace all references to one root with references to another").
		WithArg(
			clib.
				NewArg("source", "path to the source root").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithArg(
			clib.
				NewArg("replacement", "path to the replacement root").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(
			clib.NewOption(
				"include",
				"choose which roots are included in the search to find references to replace. the include filter is a CEL expression with access to a 'root' object that can be used to filter on properties of the root.",
			).WithType(clib.OptionTypeMultiString),
		).
		WithOption(
			clib.NewOption(
				"exclude",
				"choose which roots are excluded in the search to find references to replace.  the exclude filter is a CEL expression with access to a 'root' object that can be used to filter on properties of the root.",
			).WithType(clib.OptionTypeMultiString),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var source string = args[0]
			var dest string = args[1]
			var err error

			err, includeExcludeFilter := ArgRootFilterFromIncludeExclude(task.SERIAL_CONTEXT, req)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error parsing include/exclude filters")
				return 1
			}

			err, source = dydfs.PartialEvalSymlinks(task.SERIAL_CONTEXT, source)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while source root path")
				return 1
			}

			err, dest = dydfs.PartialEvalSymlinks(task.SERIAL_CONTEXT, dest)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while dest root path")
				return 1
			}

			err, garden := dryad.Garden(source).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving garden")
				return 1
			}	

			err, roots := garden.Roots().Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error resolving garden roots")
				return 1
			}

			err, safeSourceRoot := roots.Root(source).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving source root")
				return 1
			}
	
			err, safeDestRoot := roots.Root(dest).Resolve(task.SERIAL_CONTEXT)
			if err != nil {
				zlog.Fatal().Err(err).Msg("error while resolving dest root")
				return 1
			}	

			err = safeSourceRoot.Replace(
				task.SERIAL_CONTEXT,
				dryad.RootReplaceRequest{
					Filter: includeExcludeFilter,
					Dest: &safeDestRoot,
				},
			)

			if err != nil {
				zlog.Fatal().Err(err).Msg("error while replacing root")
				return 1
			}

			return 0
		})

	command = LoggingCommand(command)


	return command
}()
