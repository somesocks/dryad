package main

import (
	"os"

	cli "dryad/cli"
)

var Version string
var Fingerprint string

func main() {
	app := cli.BuildCLI(
		Version,
		Fingerprint,
	)

	// lie to cli about the name of the tool,
	// so that the help always shows the name of the command as
	// `dryad`
	args := os.Args
	args[0] = "dryad"
	os.Exit(app.Run(args, os.Stdout))
}
