package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var mkdirTemp = diagnostics.BindA2R1(
	"os.mkdir_temp",
	func(dir string, pattern string) string {
		return dir
	},
	func(dir string, pattern string) (error, string) {
		path, err := stdos.MkdirTemp(dir, pattern)
		return err, path
	},
)

var MkdirTemp = func(dir string, pattern string) (string, error) {
	err, path := mkdirTemp(dir, pattern)
	return path, err
}
