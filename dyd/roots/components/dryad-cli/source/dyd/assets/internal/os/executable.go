package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var executable = diagnostics.BindA0R1(
	"os.executable",
	func() (error, string) {
		path, err := stdos.Executable()
		return err, path
	},
)

var Executable = func() (string, error) {
	err, path := executable()
	return path, err
}
