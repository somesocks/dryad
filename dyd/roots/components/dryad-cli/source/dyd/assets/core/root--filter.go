package core

import (
	"dryad/task"
	// "fmt"
	"errors"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	zlog "github.com/rs/zerolog/log"
)

type RootFilterRequest struct {
	Expression string
}


var RootFilterCelEnv = func () *cel.Env {

	var root_path_fun = cel.Function(
		"path",
		cel.MemberOverload(
			"Root_path",
			[]*cel.Type{cel.ObjectType("Root")},
			cel.StringType,
			cel.UnaryBinding(
				func (arg ref.Val) ref.Val {		
					wrapper, ok := arg.Value().(RootCelWrapper)

					if !ok {
						return types.NewErr("invalid type for path method")
					}					

					return wrapper.Path()
				},
			),
		),
	)

	var trait_fun = cel.Function(
		"trait",
		cel.MemberOverload(
			"Root_trait",
			[]*cel.Type{cel.ObjectType("Root"), cel.StringType},
			cel.StringType,
			cel.BinaryBinding(
				func (wrapperRef ref.Val, traitRef ref.Val) ref.Val {		
					wrapper, ok := wrapperRef.Value().(RootCelWrapper)

					if !ok {
						return types.NewErr("invalid type for root")
					}					

					path, ok := traitRef.Value().(string)
					if !ok {
						return types.NewErr("invalid type for trait path")
					}					

					zlog.Trace().
						Str("root", wrapper.root.BasePath).
						Str("trait", path).
						Msg("CEL calling trait")

					err, traits := wrapper.root.Traits().Resolve(wrapper.ctx)
					if err != nil {
						if errors.Is(err, ErrorNoRootTraits) {
							return types.String("")
						}
						zlog.Error().
							Err(err).
							Msg("error getting root traits")
						return types.NullValue						
					}

					err, trait := traits.Trait(path).Resolve(wrapper.ctx)
					if err != nil {
						if errors.Is(err, ErrorNoRootTrait) {
							return types.String("")
						}
						zlog.Error().
							Err(err).
							Msg("error getting root trait")
						return types.NullValue
					}

					err, value := trait.Get(wrapper.ctx)
					if err != nil {
						zlog.Error().
							Err(err).
							Msg("error getting trait value")
						return types.NullValue
					}

					zlog.Trace().
						Str("root", wrapper.root.BasePath).
						Str("trait", path).
						Str("value", value).
						Msg("CEL trait value")

					return types.String(value)
				},
			),
		),
	)

	env, err := cel.NewEnv(
		cel.Types(cel.ObjectType("Root")),
		trait_fun,
		root_path_fun,
		cel.Variable("root", cel.ObjectType("Root")),
	)

	if err != nil {
		zlog.Error().
			Err(err).
			Msg("error generating CEL environment")
		panic(err)
		// return err, false
	}	


	return env
}();

func (root *SafeRootReference) Filter(
	ctx *task.ExecutionContext,
	req RootFilterRequest,
) (error, bool) {
	var err error

	// compile CEL expression
	ast, issues := RootFilterCelEnv.Compile(req.Expression)
	if issues != nil && issues.Err() != nil {
		zlog.Error().
			Err(err).
			Msg("error compiling CEL expression")
		return issues.Err(), false
	}

	// create CEL program
	prg, err := RootFilterCelEnv.Program(ast)
	if err != nil {
		zlog.Error().
			Err(err).
			Msg("error generating CEL program")
		return err, false
	}	

	// Evaluate with the given directory context
	wrapper := RootCelWrapper{
		root: root,
		ctx: ctx,
	}

	result, _, err := prg.Eval(map[string]any{
		"root": &wrapper,
	})
	if err != nil {
		zlog.Error().
			Err(err).
			Msg("error evaluating CEL program")
		return err, false 
	} else if result.Type() != types.BoolType {
		return errors.New("expected boolean result from filter"), false
	}
	
	zlog.Trace().
		Str("root", root.BasePath).
		Bool("result", result.Value().(bool)).
		Msg("CEL filter result")

	return nil, result.Value().(bool)
}