package core

import (
	fs2 "dryad/filesystem"
	"dryad/task"

	"dryad/internal/os"
	"net/url"
	"path/filepath"
	"strings"
	// "errors"
	// zlog "github.com/rs/zerolog/log"
)

func (rootRequirement *SafeRootRequirementReference) Replace(ctx *task.ExecutionContext, target *SafeRootReference) error {
	var err error
	var linkTarget string

	err, targetSpec := rootRequirement.TargetSpec(ctx)
	if err != nil {
		return err
	}

	err, variantSelector := variantDescriptorEncodeURL(targetSpec.VariantSelector)
	if err != nil {
		return err
	}

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
	linkUrl.RawQuery = strings.TrimPrefix(variantSelector, "?")

	err, _ = fs2.Remove(ctx, rootRequirement.BasePath)
	if err != nil {
		return err
	}

	err = os.WriteFile(rootRequirement.BasePath, []byte(linkUrl.String()), 0644)
	if err != nil {
		return err
	}

	return nil
}
