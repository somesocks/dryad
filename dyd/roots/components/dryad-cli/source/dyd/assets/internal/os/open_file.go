package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var openFile = diagnostics.BindA3R1(
	"os.open_file",
	func(name string, _ int, _ stdos.FileMode) string {
		return name
	},
	func(name string, flag int, perm stdos.FileMode) (error, *stdos.File) {
		file, err := stdos.OpenFile(name, flag, perm)
		return err, file
	},
)

var OpenFile = func(name string, flag int, perm stdos.FileMode) (*stdos.File, error) {
	err, file := openFile(name, flag, perm)
	return file, err
}
