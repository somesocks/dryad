package main

import (
	"dryad/internal/os"
	"fmt"

	cli "dryad/cli"
	"dryad/diagnostics"
)

var Version string
var Fingerprint string

func main() {
	if err := diagnostics.SetupFromEnv(); err != nil {
		fmt.Fprintln(os.Stderr, "error initializing diagnostics:", err)
		os.Exit(2)
	}

	app := cli.BuildCLI(
		Version,
		Fingerprint,
	)

	// lie to cli about the name of the tool,
	// so that the help always shows the name of the command as
	// `dryad`
	args := os.Args
	args[0] = "dryad"

	exitCode := app.Run(args, os.Stdout)
	if err := diagnostics.EmitMetricsOnExit(); err != nil {
		fmt.Fprintln(os.Stderr, "error emitting diagnostics metrics:", err)
	}
	os.Exit(exitCode)
}
