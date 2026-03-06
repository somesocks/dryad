package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var create = diagnostics.BindA1R1(
	"os.create",
	func(name string) string {
		return name
	},
	func(name string) (error, *stdos.File) {
		file, err := stdos.Create(name)
		return err, file
	},
)

var Create = func(name string) (*stdos.File, error) {
	err, file := create(name)
	return file, err
}
