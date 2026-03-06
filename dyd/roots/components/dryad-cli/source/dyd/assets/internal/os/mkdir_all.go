package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var MkdirAll = diagnostics.BindA2R0(
	"os.mkdir_all",
	func(path string, _ stdos.FileMode) string {
		return path
	},
	stdos.MkdirAll,
)
