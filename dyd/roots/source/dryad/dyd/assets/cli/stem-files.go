package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
	"os"
	"path/filepath"
	"regexp"
)

var stemFilesCommand = clib.NewCommand("files", "list the files in a stem").
	WithArg(
		clib.
			NewArg("path", "path to the stem base dir").
			AsOptional().
			WithAutoComplete(AutoCompletePath),
	).
	WithOption(clib.NewOption("exclude", "a regular expression to exclude files from the list. the regexp matches against the file path relative to the stem base directory")).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args
		var options = req.Opts

		var err error
		var matchExclude *regexp.Regexp

		if options["exclude"] != nil && options["exclude"] != "" {
			matchExclude, err = regexp.Compile(options["exclude"].(string))
			if err != nil {
				log.Fatal(err)
			}
		}

		var path string
		if len(args) > 0 {
			path = args[0]
			path, err = filepath.Abs(path)
			if err != nil {
				log.Fatal(err)
			}
		}
		if path == "" {
			path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
		}

		err = dryad.StemFiles(
			dryad.StemFilesArgs{
				BasePath:  path,
				MatchDeny: matchExclude,
			},
		)
		if err != nil {
			log.Fatal(err)
		}

		return 0
	})
