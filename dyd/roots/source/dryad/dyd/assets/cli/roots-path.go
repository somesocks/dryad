package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
)

var rootsPathCommand = clib.NewCommand("path", "return the path of the roots dir").
	WithAction(func(req clib.ActionRequest) int {
		var path, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		path, err = dryad.RootsPath(path)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println(path)

		return 0
	})
