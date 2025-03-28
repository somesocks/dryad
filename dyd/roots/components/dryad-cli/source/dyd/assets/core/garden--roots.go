package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)


func (sg *SafeGardenReference) Roots() (*UnsafeRootsReference) {
	var rootsRef = UnsafeRootsReference{
		BasePath: filepath.Join(sg.BasePath, "dyd", "roots"),
		Garden: sg,
	}
	return &rootsRef
}