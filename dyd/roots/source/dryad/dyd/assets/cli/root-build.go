package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var rootBuildCommand = clib.
	NewCommand("build", "build a specified root").
	WithArg(
		clib.
			NewArg("path", "path to the root to build").
			AsOptional().
			WithAutoComplete(AutoCompletePath),
	).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args

		var path string

		if len(args) > 0 {
			path = args[0]
		}

		if !filepath.IsAbs(path) {
			wd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path = filepath.Join(wd, path)
		}

		var rootFingerprint string
		rootFingerprint, err := dryad.RootBuild(
			dryad.BuildContext{
				RootFingerprints: map[string]string{},
			},
			path,
		)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(rootFingerprint)

		return 0
	})
