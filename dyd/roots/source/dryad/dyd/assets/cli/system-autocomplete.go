package cli

import (
	clib "dryad/cli-builder"
	"fmt"
	"strings"
)

var systemAutocomplete = clib.NewCommand("autocomplete", "print out autocomplete options based on a partial command").
	WithArg(clib.NewArg("-- args", "args to pass to the command").AsOptional()).
	WithAction(func(req clib.ActionRequest) int {
		var args = req.Args[0:]
		var results = req.App.AutoComplete(args)
		fmt.Println(strings.Join(results, " "))
		return 0
	})
