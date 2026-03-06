package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var stat = diagnostics.BindA1R1(
	"os.stat",
	func(path string) string {
		return path
	},
	func(path string) (error, stdos.FileInfo) {
		info, err := stdos.Stat(path)
		return err, info
	},
)

var Stat = func(path string) (stdos.FileInfo, error) {
	err, info := stat(path)
	return info, err
}
