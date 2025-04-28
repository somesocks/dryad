
package core

import (
	// fs2 "dryad/filesystem"
	"dryad/task"

	"os"
	"net/url"
	"path/filepath"

	// "errors"
	// zlog "github.com/rs/zerolog/log"
)

func (rootRequirement *SafeRootRequirementReference) Replace(ctx * task.ExecutionContext, target *SafeRootReference) (error) {
	var err error
	var linkTarget string

	linkTarget, err = filepath.Rel(
		filepath.Dir(rootRequirement.BasePath),
		target.BasePath)
	if err != nil {
		return err
	}
	
	var linkUrl url.URL = url.URL{
		Scheme: "root",
		Opaque: linkTarget,
	}

	err = os.WriteFile(rootRequirement.BasePath, []byte(linkUrl.String()), 0644)
	if err != nil {
		return err
	}

	return nil
}