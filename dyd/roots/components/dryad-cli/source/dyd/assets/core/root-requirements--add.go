package core

import (
	"dryad/internal/os"
	"dryad/task"
	"net/url"
	stdos "os"
	"path/filepath"
	"strings"
	// zlog "github.com/rs/zerolog/log"
)

type RootRequirementsAddRequest struct {
	Dependency                *SafeRootReference
	Alias                     string
	DependencyVariantSelector string
}

func (requirements *SafeRootRequirementsReference) Add(
	ctx *task.ExecutionContext,
	req RootRequirementsAddRequest,
) (error, *SafeRootRequirementReference) {

	var alias string = req.Alias
	var depBasePath string = req.Dependency.BasePath
	var err error

	if alias == "" {
		alias = filepath.Base(depBasePath)
	}

	err, alias = RootRequirementNormalizeName(alias)
	if err != nil {
		return err, nil
	}

	err, depSelector := variantDescriptorParseURL(req.DependencyVariantSelector)
	if err != nil {
		return err, nil
	}
	err, depSelectorRaw := variantDescriptorEncodeURL(depSelector)
	if err != nil {
		return err, nil
	}

	var requirementPath = filepath.Join(requirements.BasePath, alias)

	// make sure the roots path exists before trying to link
	err = os.MkdirAll(requirements.BasePath, stdos.ModePerm)
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
	linkUrl.RawQuery = strings.TrimPrefix(depSelectorRaw, "?")

	err = stdos.WriteFile(requirementPath, []byte(linkUrl.String()), 0644)
	if err != nil {
		return err, nil
	}

	var rootRequirementRef = SafeRootRequirementReference{
		BasePath:     requirementPath,
		Requirements: requirements,
	}
	return nil, &rootRequirementRef
}
