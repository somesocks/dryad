package core

import (
	"dryad/internal/filepath"
	"dryad/task"
	"errors"
	"fmt"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"

	zlog "github.com/rs/zerolog/log"
)

type rootCelWrapper struct {
	variant *SafeRootVariantReference
	ctx     *task.ExecutionContext
}

func (wrapper *rootCelWrapper) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return *wrapper, nil
}

func (wrapper *rootCelWrapper) ConvertToType(typeValue ref.Type) ref.Val {
	switch typeValue {
	case types.StringType:
		return types.String(fmt.Sprintf("Root{BasePath: %s}", wrapper.variant.Root.BasePath))
	}
	return types.NewErr("unsupported type conversion")
}

func (wrapper *rootCelWrapper) Equal(other ref.Val) ref.Val {
	o, ok := other.(*rootCelWrapper)
	if !ok {
		return types.False
	}

	err, left := wrapper.variant.URL()
	if err != nil {
		return types.NewErr("error resolving root variant URL")
	}
	err, right := o.variant.URL()
	if err != nil {
		return types.NewErr("error resolving root variant URL")
	}

	return types.Bool(
		wrapper.variant.Root.BasePath == o.variant.Root.BasePath &&
			left == right,
	)
}

func (wrapper *rootCelWrapper) Type() ref.Type {
	return cel.ObjectType("Root")
}

func (wrapper *rootCelWrapper) Value() any {
	return *wrapper
}

func (wrapper *rootCelWrapper) Path() ref.Val {
	relRootPath, err := filepath.Rel(wrapper.variant.Root.Roots.Garden.BasePath, wrapper.variant.Root.BasePath)
	if err != nil {
		return types.NewErr("could not resolve root path")
	}

	return types.String(relRootPath)
}

func (wrapper *rootCelWrapper) Variant() ref.Val {
	err, rendered := wrapper.variant.Filesystem()
	if err != nil {
		return types.NewErr("could not resolve root variant descriptor")
	}

	return types.String(rendered)
}

var rootFilterCelEnv = func() *cel.Env {
	rootPathFun := cel.Function(
		"path",
		cel.MemberOverload(
			"Root_path",
			[]*cel.Type{cel.ObjectType("Root")},
			cel.StringType,
			cel.UnaryBinding(
				func(arg ref.Val) ref.Val {
					wrapper, ok := arg.Value().(rootCelWrapper)
					if !ok {
						return types.NewErr("invalid type for path method")
					}

					return wrapper.Path()
				},
			),
		),
	)

	rootVariantFun := cel.Function(
		"variant",
		cel.MemberOverload(
			"Root_variant",
			[]*cel.Type{cel.ObjectType("Root")},
			cel.StringType,
			cel.UnaryBinding(
				func(arg ref.Val) ref.Val {
					wrapper, ok := arg.Value().(rootCelWrapper)
					if !ok {
						return types.NewErr("invalid type for variant method")
					}

					return wrapper.Variant()
				},
			),
		),
	)

	traitFun := cel.Function(
		"trait",
		cel.MemberOverload(
			"Root_trait",
			[]*cel.Type{cel.ObjectType("Root"), cel.StringType},
			cel.StringType,
			cel.BinaryBinding(
				func(wrapperRef ref.Val, traitRef ref.Val) ref.Val {
					wrapper, ok := wrapperRef.Value().(rootCelWrapper)
					if !ok {
						return types.NewErr("invalid type for root")
					}

					path, ok := traitRef.Value().(string)
					if !ok {
						return types.NewErr("invalid type for trait path")
					}

					zlog.Trace().
						Str("root", wrapper.variant.Root.BasePath).
						Str("variant", fmt.Sprintf("%v", wrapper.variant.Descriptor)).
						Str("trait", path).
						Msg("CEL calling trait")

					traits := wrapper.variant.Traits
					if traits == nil {
						return types.NullValue
					}

					unsafeTraits := UnsafeRootTraitsReference{
						BasePath: traits.BasePath,
						Root:     wrapper.variant.Root,
					}
					err, safeTraits := unsafeTraits.Resolve(wrapper.ctx)
					if err != nil {
						zlog.Error().
							Err(err).
							Msg("error getting root variant traits")
						return types.NewErr("error resolving root variant traits")
					} else if safeTraits == nil {
						return types.NullValue
					}

					err, trait := safeTraits.Trait(path).Resolve(wrapper.ctx)
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

					return types.String(value)
				},
			),
		),
	)

	env, err := cel.NewEnv(
		cel.Types(cel.ObjectType("Root")),
		traitFun,
		rootPathFun,
		rootVariantFun,
		cel.Variable("root", cel.ObjectType("Root")),
	)
	if err != nil {
		zlog.Error().Err(err).Msg("error generating CEL environment")
		panic(err)
	}

	return env
}()

type RootCelFilterRequest struct {
	Include []string
	Exclude []string
}

func RootVariantCelFilter(request RootCelFilterRequest) (error, RootVariantFilter) {
	includeFilters := make([]cel.Program, len(request.Include))
	excludeFilters := make([]cel.Program, len(request.Exclude))

	for k, v := range request.Include {
		ast, issues := rootFilterCelEnv.Compile(v)
		if issues != nil && issues.Err() != nil {
			zlog.Error().Err(issues.Err()).Msg("error compiling CEL expression")
			return issues.Err(), nil
		}

		prg, err := rootFilterCelEnv.Program(ast)
		if err != nil {
			zlog.Error().Err(err).Msg("error generating CEL program")
			return err, nil
		}
		includeFilters[k] = prg
	}

	for k, v := range request.Exclude {
		ast, issues := rootFilterCelEnv.Compile(v)
		if issues != nil && issues.Err() != nil {
			zlog.Error().Err(issues.Err()).Msg("error compiling CEL expression")
			return issues.Err(), nil
		}

		prg, err := rootFilterCelEnv.Program(ast)
		if err != nil {
			zlog.Error().Err(err).Msg("error generating CEL program")
			return err, nil
		}
		excludeFilters[k] = prg
	}

	return nil, func(ctx *task.ExecutionContext, variant *SafeRootVariantReference) (error, bool) {
		wrappedRef := rootCelWrapper{
			variant: variant,
			ctx:     ctx,
		}

		for _, prg := range includeFilters {
			out, _, err := prg.Eval(map[string]any{"root": &wrappedRef})
			if err != nil {
				return err, false
			}

			val, ok := out.Value().(bool)
			if !ok {
				return errors.New("non-boolean expression in include filter"), false
			}
			if !val {
				return nil, false
			}
		}

		for _, prg := range excludeFilters {
			out, _, err := prg.Eval(map[string]any{"root": &wrappedRef})
			if err != nil {
				return err, false
			}

			val, ok := out.Value().(bool)
			if !ok {
				return errors.New("non-boolean expression in exclude filter"), false
			}
			if val {
				return nil, false
			}
		}

		return nil, true
	}
}

func RootCelFilter(request RootCelFilterRequest) (error, RootFilter) {
	err, variantFilter := RootVariantCelFilter(request)
	if err != nil {
		return err, nil
	}

	return nil, RootVariantFilterToRootFilterAny(variantFilter)
}
