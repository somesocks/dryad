package core

import (
	"path/filepath"
	"dryad/task"

	"net/url"
	"os"
	// zlog "github.com/rs/zerolog/log"
)

type RootRequirementsAddRequest struct {
	Dependency *SafeRootReference
	Alias string
}

func (requirements *SafeRootRequirementsReference) Add(
	ctx * task.ExecutionContext,
	req RootRequirementsAddRequest,
) (error, *SafeRootRequirementReference) {

	var alias string = req.Alias
	var depBasePath string = req.Dependency.BasePath
	var err error

	if alias == "" {
		alias = filepath.Base(depBasePath)
	}

	var requirementPath = filepath.Join(requirements.BasePath, alias)

	// make sure the roots path exists before trying to link
	err = os.MkdirAll(requirements.BasePath, os.ModePerm)
	if err != nil {
		return err, nil
	}

	var linkPath string
	linkPath, err = filepath.Rel(requirements.BasePath, depBasePath)
	if err != nil {
		return err, nil
	}

	var linkUrl url.URL = url.URL{
		Scheme: "root",
		Opaque: linkPath,
	}

	err = os.WriteFile(requirementPath, []byte(linkUrl.String()), 0644)
	if err != nil {
		return err, nil
	}

	var rootRequirementRef = SafeRootRequirementReference{
		BasePath: requirementPath,
		Requirements: requirements,
	}
	return nil, &rootRequirementRef
}