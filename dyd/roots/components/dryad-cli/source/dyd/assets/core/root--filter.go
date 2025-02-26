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

func (root *SafeRootReference) Filter(
	ctx *task.ExecutionContext,
	req RootFilterRequest,
) (error, bool) {

	var trait_fun = cel.Function(
		"trait",
		cel.Overload(
			"trait_string",
			[]*cel.Type{cel.StringType},
			cel.StringType,
			cel.UnaryBinding(
				func (arg ref.Val) ref.Val {		
					path := arg.Value().(string)

					zlog.Trace().
						Str("root", root.BasePath).
						Str("trait", path).
						Msg("CEL calling trait")

					err, traits := root.Traits().Resolve(ctx)
					if err != nil {
						if errors.Is(err, ErrorNoTraits) {
							return types.String("")
						}
						zlog.Error().
							Err(err).
							Msg("error getting root traits")
						return types.NullValue						
					}

					err, trait := traits.Trait(path).Resolve(ctx)
					if err != nil {
						if errors.Is(err, ErrorNoTrait) {
							return types.String("")
						}
						zlog.Error().
							Err(err).
							Msg("error getting root trait")
						return types.NullValue
					}

					err, value := trait.Get(ctx)
					if err != nil {
						zlog.Error().
							Err(err).
							Msg("error getting trait value")
						return types.NullValue
					}

					zlog.Trace().
						Str("root", root.BasePath).
						Str("trait", path).
						Str("value", value).
						Msg("CEL trait value")

					return types.String(value)
				},
			),
		),
	)

	env, err := cel.NewEnv(
		trait_fun,
	)
	if err != nil {
		zlog.Error().
			Err(err).
			Msg("error generating CEL environment")
		return err, false
	}	

	// compile CEL expression
	ast, issues := env.Compile(req.Expression)
	if issues != nil && issues.Err() != nil {
		zlog.Error().
			Err(err).
			Msg("error compiling CEL expression")
		return issues.Err(), false
	}

	// create CEL program
	prg, err := env.Program(ast)
	if err != nil {
		zlog.Error().
			Err(err).
			Msg("error generating CEL program")
		return err, false
	}	

	// Evaluate with the given directory context
	result, _, err := prg.Eval(map[string]interface{}{})
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