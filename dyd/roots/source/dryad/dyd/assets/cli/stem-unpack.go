package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"log"
	"os"
)

var stemUnpackCommand = clib.NewCommand("unpack", "unpack a stem archive at the target path and import it into the current garden").
	WithArg(clib.NewArg("archive", "the path to the archive to unpack")).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args

		var stemPath = args[0]

		gardenPath, err := os.Getwd()
		if err != nil {
			log.Fatal(err)
		}

		targetPath, err := dryad.StemUnpack(gardenPath, stemPath)
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(targetPath)
		return 0
	})
