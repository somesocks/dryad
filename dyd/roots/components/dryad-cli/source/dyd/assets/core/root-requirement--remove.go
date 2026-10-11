package core

import (
	// fs2 "dryad/filesystem"
	"context"
	"dryad/internal/os"
	// "dryad/internal/filepath"
	// "errors"
	// zlog "github.com/rs/zerolog/log"
)

func (rootRequirement *SafeRootRequirementReference) Remove(ctx context.Context) error {
	var err error

	err = os.Remove(rootRequirement.BasePath)
	return err
}
