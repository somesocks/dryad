package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var WriteFile = diagnostics.BindA3R0(
	"os.write_file",
	func(name string, _ []byte, _ stdos.FileMode) string {
		return name
	},
	stdos.WriteFile,
)
