package core

import (
	"bytes"

	"text/template"
)

type pathStubRequest struct {
	BaseName    string
	CommandName string
}

var PATH_STUB_TEMPLATE, _ = template.
	New("path_stub").
	Parse(
`#!/bin/sh
set -eu
export DYD_STEM="$(dirname $0)/../dependencies/{{.BaseName}}"
export PATH="$DYD_STEM/dyd/path:$PATH"
exec "$DYD_STEM/dyd/commands/{{.CommandName}}" "$@"
`)

func rootBuild_pathStub(baseName string, commandName string) string {
	var buffer bytes.Buffer
	PATH_STUB_TEMPLATE.Execute(&buffer, pathStubRequest{
		BaseName:    baseName,
		CommandName: commandName,
	})

	return buffer.String()
}
