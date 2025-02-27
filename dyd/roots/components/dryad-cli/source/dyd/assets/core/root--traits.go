package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)

func (root *SafeRootReference) Traits() (*UnsafeRootTraitsReference) {
	var rootTraitsRef = UnsafeRootTraitsReference{
		BasePath: filepath.Join(root.BasePath, "dyd", "traits"),
		Root: root,
	}
	return &rootTraitsRef
}