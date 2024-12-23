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
		`#!/usr/bin/env sh
set -eu
STEM_PATH="$(dirname $0)/../dependencies/{{.BaseName}}"
PATH="$STEM_PATH/dyd/path:$PATH" \
DYD_STEM="$STEM_PATH" \
"$STEM_PATH/dyd/commands/{{.CommandName}}" "$@"
`)

func rootBuild_pathStub(baseName string, commandName string) string {
	var buffer bytes.Buffer
	PATH_STUB_TEMPLATE.Execute(&buffer, pathStubRequest{
		BaseName:    baseName,
		CommandName: commandName,
	})

	return buffer.String()
}
