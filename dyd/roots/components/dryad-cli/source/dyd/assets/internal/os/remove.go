package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var Remove = diagnostics.BindA1R0(
	"os.remove",
	func(path string) string {
		return path
	},
	stdos.Remove,
)
