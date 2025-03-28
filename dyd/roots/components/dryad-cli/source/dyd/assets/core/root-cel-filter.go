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


type rootCelWrapper struct {
	root *SafeRootReference
	ctx *task.ExecutionContext
}

func (wrapper *rootCelWrapper) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return *wrapper, nil
}

func (wrapper *rootCelWrapper) ConvertToType(typeValue ref.Type) ref.Val {
	switch typeValue {
	case types.StringType:
		return types.String(fmt.Sprintf("Root{BasePath: %s}", wrapper.root.BasePath))
	// case types.TypeType:
	// 	return cel.ObjectType("Root")
	}
	return types.NewErr("unsupported type conversion")
}

func (wrapper *rootCelWrapper) Equal(other ref.Val) ref.Val {
	o, ok := other.(*rootCelWrapper)
	if !ok {
		return types.False
	}
	return types.Bool(wrapper.root.BasePath == o.root.BasePath)
}

func (wrapper *rootCelWrapper) Type() ref.Type {
	return cel.ObjectType("Root")
}

func (wrapper *rootCelWrapper) Value() any {
	return *wrapper
}

func (wrapper *rootCelWrapper) Path() ref.Val {
	var relRootPath string
	var err error

	relRootPath, err = filepath.Rel(wrapper.root.Roots.Garden.BasePath, wrapper.root.BasePath)
	if err != nil {
		return types.NewErr("could not resolve root path")
	}

	return types.String(relRootPath)	
}


var rootFilterCelEnv = func () *cel.Env {

	var root_path_fun = cel.Function(
		"path",
		cel.MemberOverload(
			"Root_path",
			[]*cel.Type{cel.ObjectType("Root")},
			cel.StringType,
			cel.UnaryBinding(
				func (arg ref.Val) ref.Val {		
					wrapper, ok := arg.Value().(rootCelWrapper)

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
					wrapper, ok := wrapperRef.Value().(rootCelWrapper)

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
						zlog.Error().
							Err(err).
							Msg("error getting root traits")
						return types.NewErr("error resolving root traits")
					} else if traits == nil {
						return types.NullValue
					}

					err, trait := traits.Trait(path).Resolve(wrapper.ctx)
					if err != nil {
						zlog.Error().
							Err(err).
							Msg("error getting root trait")
						return types.NullValue
					} else if trait == nil {
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


type RootCelFilterRequest struct {
	Include []string
	Exclude [] string
}

func RootCelFilter(request RootCelFilterRequest) (error, func(ctx *task.ExecutionContext, ref *SafeRootReference) (error, bool)) {
	var includeFilters []cel.Program = make([]cel.Program, len(request.Include))
	var excludeFilters []cel.Program = make([]cel.Program, len(request.Exclude))
	var filter func(ctx *task.ExecutionContext, ref *SafeRootReference) (error, bool)

	for k, v := range request.Include {
		// compile CEL expression
		ast, issues := rootFilterCelEnv.Compile(v)
		if issues != nil && issues.Err() != nil {
			zlog.Error().
				Err(issues.Err()).
				Msg("error compiling CEL expression")
			return issues.Err(), nil
		}

		// create CEL program
		prg, err := rootFilterCelEnv.Program(ast)
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
		ast, issues := rootFilterCelEnv.Compile(v)
		if issues != nil && issues.Err() != nil {
			zlog.Error().
				Err(issues.Err()).
				Msg("error compiling CEL expression")
			return issues.Err(), nil
		}

		// create CEL program
		prg, err := rootFilterCelEnv.Program(ast)
		if err != nil {
			zlog.Error().
				Err(err).
				Msg("error generating CEL program")
			return err, nil
		}	

		excludeFilters[k] = prg
	}

	filter = func(ctx *task.ExecutionContext, root *SafeRootReference) (error, bool) {
		var matchesInclude = false
		var matchesExclude = false

		if len(includeFilters) == 0 { matchesInclude = true }

		var rootCelWrapper = rootCelWrapper{
			root: root,
			ctx: ctx,
		}
		var celArgs = map[string]any{
			"root": &rootCelWrapper,
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
