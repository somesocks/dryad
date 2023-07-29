package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
)

var stemPathCommand = clib.NewCommand("path", "return the base path of the current root").
	// WithArg(clib.NewArg("path", "path to the stem base dir")).
	WithAction(func(req clib.ActionRequest) int {
		var path, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		path, err = dryad.StemPath(path)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(path)

		return 0
	})
