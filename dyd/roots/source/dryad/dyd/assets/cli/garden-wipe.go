package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
	"os"
)

var gardenWipeCommand = clib.
	NewCommand("wipe", "clear all build artifacts out of the garden").
	WithAction(func(req clib.ActionRequest) int {
		var path, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		err = dryad.GardenWipe(
			path,
		)
		if err != nil {
			log.Fatal(err)
		}

		return 0
	})
