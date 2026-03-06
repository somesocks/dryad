package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var Link = diagnostics.BindA2R0(
	"os.link",
	func(oldPath string, newPath string) string {
		return oldPath
	},
	stdos.Link,
)
