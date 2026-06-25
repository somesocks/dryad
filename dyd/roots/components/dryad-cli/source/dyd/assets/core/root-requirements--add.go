package core

import (
	"dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/task"
	"fmt"
	"net/url"
	"strings"
	// zlog "github.com/rs/zerolog/log"
)

type RootRequirementsAddRequest struct {
	Dependency                *SafeRootReference
	Alias                     string
	DependencyVariantSelector string
}

type RootRequirementsAddEnvRequest struct {
	Alias  string
	Target string
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
	linkUrl.RawQuery = strings.TrimPrefix(depSelectorRaw, "?")

	err = os.WriteFile(requirementPath, []byte(linkUrl.String()), 0644)
	if err != nil {
		return err, nil
	}

	var rootRequirementRef = SafeRootRequirementReference{
		BasePath:     requirementPath,
		Requirements: requirements,
	}
	return nil, &rootRequirementRef
}

func (requirements *SafeRootRequirementsReference) AddEnv(
	ctx *task.ExecutionContext,
	req RootRequirementsAddEnvRequest,
) (error, *SafeRootRequirementReference) {
	err, envSpec, isEnv := rootRequirementParseEnvTarget(req.Target)
	if err != nil {
		return err, nil
	}
	if !isEnv {
		return fmt.Errorf("env requirement target must use env scheme: %s", req.Target), nil
	}

	alias := req.Alias
	if alias == "" {
		alias = envSpec.Name
	}

	err, aliasName, condition := rootRequirementParseName(alias)
	if err != nil {
		return err, nil
	}
	err, _ = rootRequirementCanonicalEnvName(aliasName)
	if err != nil {
		return err, nil
	}
	err, alias = rootRequirementEncodeName(aliasName, condition)
	if err != nil {
		return err, nil
	}

	if err := os.MkdirAll(requirements.BasePath, os.ModePerm); err != nil {
		return err, nil
	}

	requirementPath := filepath.Join(requirements.BasePath, alias)
	if err := os.WriteFile(requirementPath, []byte(rootRequirementEnvTargetString(envSpec.Name, envSpec.Fingerprint)), 0644); err != nil {
		return err, nil
	}

	return nil, &SafeRootRequirementReference{
		BasePath:     requirementPath,
		Requirements: requirements,
	}
}
