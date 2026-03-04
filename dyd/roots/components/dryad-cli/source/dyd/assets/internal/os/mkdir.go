package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var Mkdir = diagnostics.BindA2R0(
	"os.mkdir",
	func(path string, _ stdos.FileMode) string {
		return path
	},
	stdos.Mkdir,
)
