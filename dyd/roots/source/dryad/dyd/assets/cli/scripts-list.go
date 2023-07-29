package cli

import (
	clib "dryad/cli-builder"
	dryad "dryad/core"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sort"
	"strings"
)

var scriptsListAction = func(req clib.ActionRequest) int {
	var options = req.Opts

	basePath, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	var scope string
	if options["scope"] != nil {
		scope = options["scope"].(string)
	} else {
		var err error
		scope, err = dryad.ScopeGetDefault(scope)
		fmt.Println("[info] loading default scope:", scope)
		if err != nil {
			log.Fatal(err)
		}
	}

	var showPath bool

	if options["path"] != nil {
		showPath = options["path"].(bool)
	} else {
		showPath = false
	}

	// if the scope is unset, bypass expansion and run the action directly
	if scope == "" || scope == "none" {
		log.Fatal("no scope set, can't find command")
	} else {
		fmt.Println("[info] using scope:", scope)
	}

	var scripts []string

	err = dryad.ScriptsWalk(dryad.ScriptsWalkRequest{
		BasePath: basePath,
		Scope:    scope,
		OnMatch: func(path string, info fs.FileInfo) error {
			if showPath {
				scripts = append(scripts, path)
			} else {
				var name string = info.Name()
				var script string = "dryad script run " + strings.TrimPrefix(name, "script-run-")
				scripts = append(scripts, script)
			}
			return nil
		},
	})
	if err != nil {
		log.Fatal(err)
	}

	sort.Strings(scripts)

	for _, script := range scripts {
		fmt.Println(script)
	}

	return 0
}

var scriptsListCommand = clib.NewCommand("list", "list all available scripts in the current scope").
	WithOption(clib.NewOption("scope", "set the scope for the command")).
	WithOption(clib.NewOption("path", "print the path to the scripts instead of the script run command").WithType(clib.OptionTypeBool)).
	WithAction(scopeHandler(scriptsListAction))
