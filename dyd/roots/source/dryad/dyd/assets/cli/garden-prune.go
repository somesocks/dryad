package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"log"
	"os"
)

var gardenPruneCommand = clib.NewCommand("prune", "clear all build artifacts out of the garden not actively linked to a sprout or a root").
	WithAction(func(req clib.ActionRequest) int {
		var path, err = os.Getwd()
		if err != nil {
			log.Fatal(err)
		}
		err = dryad.GardenPrune(
			path,
		)
		if err != nil {
			log.Fatal(err)
		}

		return 0
	})
