package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)

func (root *SafeRootReference) Requirements() (*UnsafeRootRequirementsReference) {
	var rootRequirementsRef = UnsafeRootRequirementsReference{
		BasePath: filepath.Join(root.BasePath, "dyd", "requirements"),
		Root: root,
	}
	return &rootRequirementsRef
}