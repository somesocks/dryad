package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
	"regexp"
)

var stemFingerprintCommand = clib.NewCommand("fingerprint", "calculate the fingerprint for a stem dir").
	WithArg(
		clib.
			NewArg("path", "path to the stem base dir").
			AsOptional().
			WithAutoComplete(ArgAutoCompletePath),
	).
	WithOption(clib.NewOption("exclude", "a regular expression to exclude files from the fingerprint calculation. the regexp matches against the file path relative to the stem base directory")).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args
		var options = req.Opts

		var err error
		var matchExclude *regexp.Regexp

		if options["exclude"] != "" {
			matchExclude, err = regexp.Compile(options["exclude"].(string))
			if err != nil {
				log.Fatal(err)
			}
		}

		var path string
		if len(args) > 0 {
			path = args[0]
		}
		if path == "" {
			path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
		}

		var fingerprintString, fingerprintErr = dryad.StemFingerprint(
			dryad.StemFingerprintArgs{
				BasePath:  path,
				MatchDeny: matchExclude,
			},
		)
		if fingerprintErr != nil {
			log.Fatal(fingerprintErr)
		}
		fmt.Println(fingerprintString)

		return 0
	})
