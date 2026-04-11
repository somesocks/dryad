package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	dydfs "dryad/filesystem"
	"dryad/task"
	"fmt"
	"net/url"

	zlog "github.com/rs/zerolog/log"
)

var rootReplaceCommand = func() clib.Command {
	type ParsedArgs struct {
		SourcePath           string
		SourceSelector       dryad.VariantDescriptor
		SourceHasSelector    bool
		DestPath             string
		DestSelector         dryad.VariantDescriptor
		DestHasSelector      bool
		Parallel             int
		FromStdinFilter      dryad.RootVariantFilter
		IncludeExcludeFilter dryad.RootVariantFilter
	}

	var parseRootReplaceTargetRef = func(raw string) (error, parsedRootRef) {
		targetURL, err := url.Parse(raw)
		if err != nil {
			return err, parsedRootRef{}
		}

		if targetURL.Scheme == "" {
			return parseRootRef(raw)
		}

		if targetURL.Scheme != "root" {
			return fmt.Errorf("unsupported scheme for root ref: %s", targetURL.Scheme), parsedRootRef{}
		}

		if targetURL.Fragment != "" {
			return fmt.Errorf("variant descriptor fragments are not supported; use query parameters with '&'"), parsedRootRef{}
		}

		targetPath := targetURL.Path
		if targetURL.Opaque != "" {
			targetPath = targetURL.Opaque
		}
		if targetPath == "" {
			return fmt.Errorf("missing root ref path"), parsedRootRef{}
		}

		selector := dryad.VariantDescriptor{}
		hasSelector := targetURL.RawQuery != ""
		if hasSelector {
			err, variantContext := dryad.RootVariantContextFromURL("?" + targetURL.RawQuery)
			if err != nil {
				return err, parsedRootRef{}
			}
			selector = variantContext.Descriptor
		}

		return nil, parsedRootRef{
			Path:        targetPath,
			Selector:    selector,
			HasSelector: hasSelector,
		}
	}

	var parseArgs = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts

		var sourceRaw string = args[0]
		var destRaw string = args[1]
		var source string
		var dest string
		var err error
		var parallel int

		if options["parallel"] != nil {
			parallel = int(options["parallel"].(int64))
		} else {
			parallel = PARALLEL_COUNT_DEFAULT
		}

		err, includeExcludeFilter := ArgRootVariantFilterFromIncludeExclude(ctx, req)
		if err != nil {
			return err, ParsedArgs{}
		}

		err, fromStdinFilter := ArgRootVariantFilterFromStdin(ctx, req)
		if err != nil {
			return err, ParsedArgs{}
		}

		err, sourceRef := parseRootReplaceTargetRef(sourceRaw)
		if err != nil {
			return err, ParsedArgs{}
		}

		err, source = dydfs.PartialEvalSymlinks(ctx, sourceRef.Path)
		if err != nil {
			return err, ParsedArgs{}
		}

		err, destRef := parseRootReplaceTargetRef(destRaw)
		if err != nil {
			return err, ParsedArgs{}
		}

		err, dest = dydfs.PartialEvalSymlinks(ctx, destRef.Path)
		if err != nil {
			return err, ParsedArgs{}
		}

		return nil, ParsedArgs{
			SourcePath:           source,
			SourceSelector:       sourceRef.Selector,
			SourceHasSelector:    sourceRef.HasSelector,
			DestPath:             dest,
			DestSelector:         destRef.Selector,
			DestHasSelector:      destRef.HasSelector,
			Parallel:             parallel,
			FromStdinFilter:      fromStdinFilter,
			IncludeExcludeFilter: includeExcludeFilter,
		}
	}

	var replaceRoot = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
		err, garden := dryad.Garden(args.SourcePath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, roots := garden.Roots().Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, safeSourceRoot := roots.Root(args.SourcePath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err, safeDestRoot := roots.Root(args.DestPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		err = safeSourceRoot.Replace(
			ctx,
			dryad.RootReplaceRequest{
				Filter: dryad.RootVariantFiltersCompose(
					args.FromStdinFilter,
					args.IncludeExcludeFilter,
				),
				Source: dryad.RootReplaceTargetSpec{
					Root:               &safeSourceRoot,
					VariantSelector:    args.SourceSelector,
					HasVariantSelector: args.SourceHasSelector,
				},
				Dest: dryad.RootReplaceTargetSpec{
					Root:               &safeDestRoot,
					VariantSelector:    args.DestSelector,
					HasVariantSelector: args.DestHasSelector,
				},
			},
		)

		return err, nil
	}

	replaceRoot = task.WithContext(
		replaceRoot,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			replaceRoot,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Error().Err(err).Msg("error while replacing root")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("replace", "replace matching root requirement target refs with another target ref").
		WithArg(
			clib.
				NewArg("old_target", "root path or root ref to match in requirements (for example ../dep, ../dep~os=linux, or root:../dep?os=linux)").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithArg(
			clib.
				NewArg("new_target", "root path or root ref patch to rewrite matching requirements to (for example ../dep, ../dep~os=linux, or root:../dep?os=linux)").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(
			clib.NewOption(
				"include",
				"choose which root variants are included in the search to find references to replace. the include filter is a CEL expression with access to a 'root' object for each root variant.",
			).WithType(clib.OptionTypeMultiString),
		).
		WithOption(
			clib.NewOption(
				"exclude",
				"choose which root variants are excluded in the search to find references to replace. the exclude filter is a CEL expression with access to a 'root' object for each root variant.",
			).WithType(clib.OptionTypeMultiString),
		).
		WithOption(
			clib.NewOption(
				"from-stdin",
				"if set, read a list of root refs from stdin to use as a base list, instead of all root variants. include and exclude filters will be applied to this list. default false",
			).
				WithType(clib.OptionTypeBool),
		).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
