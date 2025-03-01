package core

import (
	"dryad/task"

	"path/filepath"
	"fmt"

	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

type SproutCelWrapper struct {
	sprout *SafeSproutReference
	ctx *task.ExecutionContext
}

func (wrapper *SproutCelWrapper) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return *wrapper, nil
}

func (wrapper *SproutCelWrapper) ConvertToType(typeValue ref.Type) ref.Val {
	switch typeValue {
	case types.StringType:
		return types.String(fmt.Sprintf("Sprout{BasePath: %s}", wrapper.sprout.BasePath))
	// case types.TypeType:
	// 	return cel.ObjectType("Sprout")
	}
	return types.NewErr("unsupported type conversion")
}

func (wrapper *SproutCelWrapper) Equal(other ref.Val) ref.Val {
	o, ok := other.(*SproutCelWrapper)
	if !ok {
		return types.False
	}
	return types.Bool(wrapper.sprout.BasePath == o.sprout.BasePath)
}

func (wrapper *SproutCelWrapper) Type() ref.Type {
	return cel.ObjectType("Sprout")
}

func (wrapper *SproutCelWrapper) Value() any {
	return *wrapper
}

func (wrapper *SproutCelWrapper) Path() ref.Val {
	var relSproutPath string
	var err error

	relSproutPath, err = filepath.Rel(wrapper.sprout.Sprouts.Garden.BasePath, wrapper.sprout.BasePath)
	if err != nil {
		return types.NewErr("could not resolve sprout path")
	}

	return types.String(relSproutPath)	
}
