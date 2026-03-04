package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var getwd = diagnostics.BindA0R1(
	"os.getwd",
	func() (error, string) {
		wd, err := stdos.Getwd()
		return err, wd
	},
)

var Getwd = func() (string, error) {
	err, wd := getwd()
	return wd, err
}
