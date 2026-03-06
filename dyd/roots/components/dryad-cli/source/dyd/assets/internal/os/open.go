package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var open = diagnostics.BindA1R1(
	"os.open",
	func(name string) string {
		return name
	},
	func(name string) (error, *stdos.File) {
		file, err := stdos.Open(name)
		return err, file
	},
)

var Open = func(name string) (*stdos.File, error) {
	err, file := open(name)
	return file, err
}
