package cli

import (
	clib "dryad/cli-builder"
	"fmt"
	"runtime"
)

var versionCommand = func(
	Version string,
	Fingerprint string,
) clib.Command {
	command := clib.NewCommand("version", "print out detailed version info").
		WithAction(func(req clib.ActionRequest) int {
			fmt.Println("version=" + Version)
			fmt.Println("source_fingerprint=" + Fingerprint)
			fmt.Println("arch=" + runtime.GOARCH)
			fmt.Println("os=" + runtime.GOOS)
			return 0
		})

	command = HelpCommand(command)

	return command
}
