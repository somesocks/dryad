package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var Rename = diagnostics.BindA2R0(
	"os.rename",
	func(oldPath string, newPath string) string {
		return oldPath
	},
	stdos.Rename,
)
