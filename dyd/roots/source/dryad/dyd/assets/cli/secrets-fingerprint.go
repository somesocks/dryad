package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var secretsFingerprintCommand = func() clib.Command {
	command := clib.NewCommand("fingerprint", "calculate the fingerprint for the secrets in a stem/root").
		WithArg(
			clib.
				NewArg("path", "path to the stem base dir").
				WithAutoComplete(ArgAutoCompletePath),
		).
		WithAction(func(req clib.ActionRequest) int {
			var args = req.Args

			var err error
			var path string

			if len(args) > 0 {
				path = args[0]
				path, err = filepath.Abs(path)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				path, err = os.Getwd()
				if err != nil {
					log.Fatal(err)
				}
			}

			// normalize the path to point to the closest secrets
			path, err = dryad.SecretsPath(path)
			if err != nil {
				log.Fatal(err)
			}

			fingerprint, err := dryad.SecretsFingerprint(
				dryad.SecretsFingerprintArgs{
					BasePath: path,
				},
			)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(fingerprint)

			return 0
		})

	command = LoggingCommand(command)
	command = HelpCommand(command)

	return command
}()
