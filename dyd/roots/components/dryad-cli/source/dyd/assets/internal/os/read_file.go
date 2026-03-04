package os

import (
	"dryad/diagnostics"
	stdos "os"
)

var readFile = diagnostics.BindA1R1(
	"os.read_file",
	func(name string) string {
		return name
	},
	func(name string) (error, []byte) {
		content, err := stdos.ReadFile(name)
		return err, content
	},
)

var ReadFile = func(name string) ([]byte, error) {
	err, content := readFile(name)
	return content, err
}
