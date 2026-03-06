package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var lstat = diagnostics.BindA1R1(
	"os.lstat",
	func(path string) string {
		return path
	},
	func(path string) (error, stdos.FileInfo) {
		info, err := stdos.Lstat(path)
		return err, info
	},
)

var Lstat = func(path string) (stdos.FileInfo, error) {
	err, info := lstat(path)
	return info, err
}
