package core

import (
	"dryad/task"
	"fmt"
	"reflect"
	"path/filepath"
	"errors"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	zlog "github.com/rs/zerolog/log"
)


type sproutCelWrapper struct {
	sprout *SafeSproutReference
	ctx *task.ExecutionContext
}

func (wrapper *sproutCelWrapper) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return *wrapper, nil
}

func (wrapper *sproutCelWrapper) ConvertToType(typeValue ref.Type) ref.Val {
	switch typeValue {
	case types.StringType:
		return types.String(fmt.Sprintf("Sprout{BasePath: %s}", wrapper.sprout.BasePath))
	// case types.TypeType:
	// 	return cel.ObjectType("Sprout")
	}
	return types.NewErr("unsupported type conversion")
}

func (wrapper *sproutCelWrapper) Equal(other ref.Val) ref.Val {
	o, ok := other.(*sproutCelWrapper)
	if !ok {
		return types.False
	}
	return types.Bool(wrapper.sprout.BasePath == o.sprout.BasePath)
}

func (wrapper *sproutCelWrapper) Type() ref.Type {
	return cel.ObjectType("Sprout")
}

func (wrapper *sproutCelWrapper) Value() any {
	return *wrapper
}

func (wrapper *sproutCelWrapper) Path() ref.Val {
	var relSproutPath string
	var err error

	relSproutPath, err = filepath.Rel(wrapper.sprout.Sprouts.Garden.BasePath, wrapper.sprout.BasePath)
	if err != nil {
		return types.NewErr("could not resolve sprout path")
	}

	return types.String(relSproutPath)	
}


var sproutFilterCelEnv = func () *cel.Env {

	var sprout_path_fun = cel.Function(
		"path",
		cel.MemberOverload(
			"Sprout_path",
			[]*cel.Type{cel.ObjectType("Sprout")},
			cel.StringType,
			cel.UnaryBinding(
				func (arg ref.Val) ref.Val {		
					wrapper, ok := arg.Value().(sproutCelWrapper)

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
			"Sprout_trait",
			[]*cel.Type{cel.ObjectType("Sprout"), cel.StringType},
			cel.StringType,
			cel.BinaryBinding(
				func (wrapperRef ref.Val, traitRef ref.Val) ref.Val {		
					wrapper, ok := wrapperRef.Value().(sproutCelWrapper)

					if !ok {
						return types.NewErr("invalid type for sprout")
					}					

					path, ok := traitRef.Value().(string)
					if !ok {
						return types.NewErr("invalid type for trait path")
					}					

					zlog.Trace().
						Str("sprout", wrapper.sprout.BasePath).
						Str("trait", path).
						Msg("CEL calling trait")

					err, traits := wrapper.sprout.Traits().Resolve(wrapper.ctx)
					if err != nil {
						if errors.Is(err, ErrorNoSproutTraits) {
							return types.String("")
						}
						zlog.Error().
							Err(err).
							Msg("error getting sprout traits")
						return types.NullValue						
					}

					err, trait := traits.Trait(path).Resolve(wrapper.ctx)
					if err != nil {
						if errors.Is(err, ErrorNoSproutTrait) {
							return types.String("")
						}
						zlog.Error().
							Err(err).
							Msg("error getting sprout trait")
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
						Str("sprout", wrapper.sprout.BasePath).
						Str("trait", path).
						Str("value", value).
						Msg("CEL trait value")

					return types.String(value)
				},
			),
		),
	)

	env, err := cel.NewEnv(
		cel.Types(cel.ObjectType("Sprout")),
		trait_fun,
		sprout_path_fun,
		cel.Variable("sprout", cel.ObjectType("Sprout")),
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


type SproutCelFilterRequest struct {
	Include []string
	Exclude [] string
}

func SproutCelFilter(request SproutCelFilterRequest) (error, func(ctx *task.ExecutionContext, ref *SafeSproutReference) (error, bool)) {
	var includeFilters []cel.Program = make([]cel.Program, len(request.Include))
	var excludeFilters []cel.Program = make([]cel.Program, len(request.Exclude))
	var filter func(ctx *task.ExecutionContext, ref *SafeSproutReference) (error, bool)

	for k, v := range request.Include {
		// compile CEL expression
		ast, issues := sproutFilterCelEnv.Compile(v)
		if issues != nil && issues.Err() != nil {
			zlog.Error().
				Err(issues.Err()).
				Msg("error compiling CEL expression")
			return issues.Err(), nil
		}

		// create CEL program
		prg, err := sproutFilterCelEnv.Program(ast)
		if err != nil {
			zlog.Error().
				Err(err).
				Msg("error generating CEL program")
			return err, nil
		}	

		includeFilters[k] = prg
	}

	for k, v := range request.Exclude {
		// compile CEL expression
		ast, issues := sproutFilterCelEnv.Compile(v)
		if issues != nil && issues.Err() != nil {
			zlog.Error().
				Err(issues.Err()).
				Msg("error compiling CEL expression")
			return issues.Err(), nil
		}

		// create CEL program
		prg, err := sproutFilterCelEnv.Program(ast)
		if err != nil {
			zlog.Error().
				Err(err).
				Msg("error generating CEL program")
			return err, nil
		}	

		excludeFilters[k] = prg
	}

	filter = func(ctx *task.ExecutionContext, sprout *SafeSproutReference) (error, bool) {
		var matchesInclude = false
		var matchesExclude = false

		if len(includeFilters) == 0 { matchesInclude = true }

		var sproutCelWrapper = sproutCelWrapper{
			sprout: sprout,
			ctx: ctx,
		}
		var celArgs = map[string]any{
			"sprout": &sproutCelWrapper,
		}
	
		for _, include := range includeFilters {
			var matchesFilter bool
			result, _, err := include.Eval(celArgs)
			if err != nil {
				zlog.Error().
					Err(err).
					Msg("error evaluating CEL filter")
				return err, false 
			} else if result.Type() != types.BoolType {
				return errors.New("expected boolean result from filter"), false
			}
			
			matchesFilter = result.Value().(bool)

			matchesInclude = matchesInclude || matchesFilter
			if matchesInclude {
				break
			}
		}

		for _, exclude := range excludeFilters {
			var matchesFilter bool
			result, _, err := exclude.Eval(celArgs)
			if err != nil {
				zlog.Error().
					Err(err).
					Msg("error evaluating CEL filter")
				return err, false 
			} else if result.Type() != types.BoolType {
				return errors.New("expected boolean result from filter"), false
			}
			
			matchesFilter = result.Value().(bool)

			matchesExclude = matchesExclude || matchesFilter
			if matchesExclude {
				break
			}
		}

		return nil, matchesInclude && !matchesExclude
	}


	return nil, filter
}
