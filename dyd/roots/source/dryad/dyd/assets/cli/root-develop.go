package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

var rootDevelopCommand = clib.NewCommand("develop", "create a temporary development environment for a root").
	WithArg(
		clib.
			NewArg("path", "path to the root to develop").
			AsOptional().
			WithAutoComplete(AutoCompletePath),
	).
	WithOption(clib.NewOption("editor", "choose the editor to run in the root development environment").WithType(clib.OptionTypeString)).
	WithOption(clib.NewOption("arg", "argument to pass to the editor").WithType(clib.OptionTypeMultiString)).
	WithOption(clib.NewOption("inherit", "inherit env variables from the host environment").WithType(clib.OptionTypeBool)).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args
		var opts = req.Opts

		var path string
		var editor string
		var editorArgs []string
		var inherit bool

		if len(args) > 0 {
			path = args[0]
		}

		if opts["editor"] != nil {
			editor = opts["editor"].(string)
		} else {
			editor = "sh"
		}

		if opts["arg"] != nil {
			editorArgs = opts["arg"].([]string)
		}

		if opts["inherit"] != nil {
			inherit = opts["inherit"].(bool)
		}

		if !filepath.IsAbs(path) {
			wd, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path = filepath.Join(wd, path)
		}

		var rootFingerprint string
		rootFingerprint, err := dryad.RootDevelop(
			dryad.BuildContext{
				RootFingerprints: map[string]string{},
			},
			path,
			editor,
			editorArgs,
			inherit,
		)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(rootFingerprint)

		return 0
	})
