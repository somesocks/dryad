package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var readlink = diagnostics.BindA1R1(
	"os.readlink",
	func(path string) string {
		return path
	},
	func(path string) (error, string) {
		target, err := stdos.Readlink(path)
		return err, target
	},
)

var Readlink = func(path string) (string, error) {
	err, target := readlink(path)
	return target, err
}
