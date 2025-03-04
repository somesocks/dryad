package core

import (
	"path/filepath"
	// "dryad/task"

	// zlog "github.com/rs/zerolog/log"
)

func (requirements *SafeRootRequirementsReference) Requirement(path string) (*UnsafeRootRequirementReference) {
	var rootRequirementRef = UnsafeRootRequirementReference{
		BasePath: filepath.Join(requirements.BasePath, path),
		Requirements: requirements,
	}
	return &rootRequirementRef
}