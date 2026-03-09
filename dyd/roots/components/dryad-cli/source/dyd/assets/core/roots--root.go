package core

import (
// "dryad/internal/filepath"
// "dryad/task"

// zlog "github.com/rs/zerolog/log"
)

func (roots *SafeRootsReference) Root(path string) *UnsafeRootReference {
	var rootRef = UnsafeRootReference{
		BasePath: path,
		Roots:    roots,
	}
	return &rootRef
}
