package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	dydfs "dryad/filesystem"
	"dryad/internal/filepath"
	"dryad/task"
	"fmt"
	"sort"

	zlog "github.com/rs/zerolog/log"
)

var rootAncestorsCommand = func() clib.Command {

	type ParsedArgs struct {
		RootPath    string
		Selector    dryad.VariantDescriptor
		HasSelector bool
		Relative    bool
		Parallel    int
	}

	var parseArgs task.Task[clib.ActionRequest, ParsedArgs] = func(ctx *task.ExecutionContext, req clib.ActionRequest) (error, ParsedArgs) {
		var args = req.Args
		var options = req.Opts
		var err error

		var rootRefRaw string
		var rootPath string
		var selector dryad.VariantDescriptor
		var hasSelector bool

		if len(args) > 0 {
			rootRefRaw = args[0]
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
			parallel = PARALLEL_COUNT_DEFAULT
		}

		err, rootRef := parseRootRef(rootRefRaw)
		if err != nil {
			return err, ParsedArgs{}
		}
		rootPath = rootRef.Path
		selector = rootRef.Selector
		hasSelector = rootRef.HasSelector

		if options["variant"] != nil {
			if hasSelector {
				return fmt.Errorf("root ancestor selector specified in both root_ref and --variant"), ParsedArgs{}
			}

			err, variantContext := dryad.RootVariantContextFromFilesystem(options["variant"].(string))
			if err != nil {
				return err, ParsedArgs{}
			}
			selector = variantContext.Descriptor
			hasSelector = true
		}

		err, rootPath = dydfs.PartialEvalSymlinks(ctx, rootPath)
		if err != nil {
			return err, ParsedArgs{}
		}

		return nil, ParsedArgs{
			RootPath:    rootPath,
			Selector:    selector,
			HasSelector: hasSelector,
			Relative:    relative,
			Parallel:    parallel,
		}
	}

	var findAncestors = func(ctx *task.ExecutionContext, args ParsedArgs) (error, any) {
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

		err, startVariants := root.ResolveBuildVariants(
			ctx,
			dryad.RootResolveBuildVariantsRequest{
				Selector: args.Selector,
			},
		)
		if err != nil {
			return err, nil
		}
		if len(startVariants) == 0 {
			return fmt.Errorf("resolved root ancestor variants are empty"), nil
		}

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

		startNodes := make([]string, 0, len(startVariants))
		startNodeSet := make(map[string]bool, len(startVariants))
		for _, variant := range startVariants {
			nodePath := rootPath

			err, variantSelectorRaw := (dryad.RootVariantContext{Descriptor: variant}).URL()
			if err != nil {
				return err, nil
			}

			node := nodePath + variantSelectorRaw
			startNodes = append(startNodes, node)
			startNodeSet[node] = true
		}

		ancestors := graph.DescendantNodes(make(dryad.TStringSet), startNodes).ToArray([]string{})
		sort.Strings(ancestors)

		for _, v := range ancestors {
			if startNodeSet[v] {
				continue
			}
			fmt.Println(v)
		}

		return nil, nil
	}

	findAncestors = task.WithContext(
		findAncestors,
		func(ctx *task.ExecutionContext, args ParsedArgs) (error, *task.ExecutionContext) {
			return nil, task.NewContext(args.Parallel)
		},
	)

	var action = task.Return(
		task.Series2(
			parseArgs,
			findAncestors,
		),
		func(err error, val any) int {
			if err != nil {
				zlog.Error().Err(err).Msg("error while finding root ancestors")
				return 1
			}

			return 0
		},
	)

	command := clib.NewCommand("ancestors", "list all package variants the selected root depends on (directly and indirectly)").
		WithArg(
			clib.
				NewArg("root_ref", "path to the root, optionally qualified with a variant selector").
				AsOptional().
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithOption(clib.NewOption("variant", "select root variants to start from (filesystem form: dimension=option+dimension=option). supports none/any/host; inherit is invalid. may resolve to multiple concrete variants").WithType(clib.OptionTypeString)).
		WithOption(clib.NewOption("relative", "print roots relative to the base garden path. default true").WithType(clib.OptionTypeBool)).
		WithAction(action)

	command = ParallelCommand(command)
	command = ScopedCommand(command)
	command = LoggingCommand(command)

	return command
}()
