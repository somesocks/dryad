
package core

import (
	// fs2 "dryad/filesystem"
	"dryad/task"

	"os"
	"path/filepath"

	"net/url"

	"fmt"
	// zlog "github.com/rs/zerolog/log"
)

func (rootRequirement *SafeRootRequirementReference) Target(ctx * task.ExecutionContext) (error, *SafeRootReference) {
	var err error
	var safeRef SafeRootReference


	linkInfo, err := os.Lstat(rootRequirement.BasePath)
	if err != nil {
		return err, nil
	}

	isSymlink := linkInfo.Mode()&os.ModeSymlink == os.ModeSymlink

	// compatibility with legacy gardens - if a requirement is a symlink
	// resolve it directly
	if isSymlink {
		linkPath, err := os.Readlink(rootRequirement.BasePath)

		// convert relative links to an absolute path
		if !filepath.IsAbs(linkPath) {
			linkPath = filepath.Join(
				filepath.Dir(rootRequirement.BasePath),
				linkPath,
			)
		}

		err, safeRef = rootRequirement.Requirements.Root.Roots.Root(linkPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		return nil, &safeRef
	} else {
		linkBytes, err := os.ReadFile(rootRequirement.BasePath)
		if err != nil {
			return err, nil
		}

		linkString := string(linkBytes)
		linkUrl, err := url.Parse(linkString)
		if err != nil {
			return err, nil
		}

		linkScheme := linkUrl.Scheme

		if linkScheme != "root" {
			return fmt.Errorf("unsupported scheme for root requirement: %s", linkScheme), nil
		}

		linkPath := linkUrl.Path
		if filepath.IsAbs(linkPath) {
			return fmt.Errorf("root requirement path must be relative", linkScheme), nil
		}

		linkPath = filepath.Join(
			filepath.Dir(rootRequirement.BasePath),
			linkPath,
		)

		err, safeRef = rootRequirement.Requirements.Root.Roots.Root(linkPath).Resolve(ctx)
		if err != nil {
			return err, nil
		}

		return nil, &safeRef
	}

}