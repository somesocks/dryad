package core

import (
	// "dryad/task"

	"path/filepath"
	"fmt"

	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

type RootCelWrapper struct {
	root *SafeRootReference
}

func (wrapper *RootCelWrapper) ConvertToNative(typeDesc reflect.Type) (any, error) {
	return *wrapper, nil
}

func (wrapper *RootCelWrapper) ConvertToType(typeValue ref.Type) ref.Val {
	switch typeValue {
	case types.StringType:
		return types.String(fmt.Sprintf("Root{BasePath: %s}", wrapper.root.BasePath))
	// case types.TypeType:
	// 	return cel.ObjectType("Root")
	}
	return types.NewErr("unsupported type conversion")
}

func (wrapper *RootCelWrapper) Equal(other ref.Val) ref.Val {
	o, ok := other.(*RootCelWrapper)
	if !ok {
		return types.False
	}
	return types.Bool(wrapper.root.BasePath == o.root.BasePath)
}

func (wrapper *RootCelWrapper) Type() ref.Type {
	return cel.ObjectType("Root")
}

func (wrapper *RootCelWrapper) Value() any {
	return *wrapper
}

func (wrapper *RootCelWrapper) Path() ref.Val {
	var relRootPath string
	var err error

	relRootPath, err = filepath.Rel(wrapper.root.Roots.Garden.BasePath, wrapper.root.BasePath)
	if err != nil {
		return types.NewErr("could not resolve root path")
	}

	return types.String(relRootPath)	
}
