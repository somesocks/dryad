package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var createTemp = diagnostics.BindA2R1(
	"os.create_temp",
	func(dir string, pattern string) string {
		return dir
	},
	func(dir string, pattern string) (error, *stdos.File) {
		file, err := stdos.CreateTemp(dir, pattern)
		return err, file
	},
)

var CreateTemp = func(dir string, pattern string) (*stdos.File, error) {
	err, file := createTemp(dir, pattern)
	return file, err
}
