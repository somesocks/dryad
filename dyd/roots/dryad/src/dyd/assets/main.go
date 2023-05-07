package main

import (
	"bufio"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	dryad "dryad/core"

	cli "dryad/cli"
)

var Version string
var Fingerprint string

func _buildCLI() cli.App {

	var app = cli.New("dryad package manager " + Version)

	var _scopeHandler = func(
		action func(req cli.ActionRequest) int,
	) func(req cli.ActionRequest) int {
		wrapper := func(req cli.ActionRequest) int {
			invocation := req.Invocation
			options := req.Opts

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

			// if the scope is unset, bypass expansion and run the action directly
			if scope == "" || scope == "none" {
				return action(req)
			} else {
				fmt.Println("[info] using scope:", scope)
			}

			settingName := strings.Join(invocation[1:], "-")

			path, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			setting, err := dryad.ScopeSettingGet(path, scope, settingName)
			if err != nil {
				log.Fatal(err)
			}

			settings, err := dryad.ScopeSettingParseShell(setting)
			if err != nil {
				log.Fatal(err)
			}

			// if the scope setting is unset,
			// there`s no need to modify the request
			if len(settings) == 0 {
				return action(req)
			}

			argsRewrite := make([]string, 0)
			index := 0

			// copy all of the arguments before the scope arg
			for index < len(os.Args) {
				var element = os.Args[index]
				index++ // do this here so it increments before the break
				if strings.HasPrefix(element, "--scope=") {
					break
				} else {
					argsRewrite = append(argsRewrite, element)
				}
			}

			// insert a null scope arg, to stop a loop
			argsRewrite = append(argsRewrite, "--scope=none")

			// insert the new args from the settings in place of the scope
			argsRewrite = append(argsRewrite, settings...)

			// copy all of the arguments after the scope arg
			for index < len(os.Args) {
				var element = os.Args[index]
				argsRewrite = append(argsRewrite, element)
				index++
			}

			fmt.Println("[info] rewriting args to:", argsRewrite)
			return app.Run(argsRewrite, os.Stdout)
		}

		return wrapper
	}

	var gardenInit = cli.NewCommand("init", "initialize a garden").
		WithArg(cli.NewArg("path", "the target path at which to initialize the garden").AsOptional()).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var path string
			var err error

			if len(args) > 0 {
				path = args[0]
			}

			err = dryad.GardenInit(path)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var gardenPath = cli.NewCommand("path", "return the base path for a garden").
		WithArg(cli.NewArg("path", "the target path at which to start for the base garden path").AsOptional()).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var path string
			var err error

			if len(args) > 0 {
				path = args[0]
			}

			path, err = dryad.GardenPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)

			return 0
		})

	var gardenBuild = cli.NewCommand("build", "build all roots in the garden").
		WithArg(cli.NewArg("path", "the target path for the garden to build").AsOptional()).
		WithOption(cli.NewOption("include", "choose which roots are included in the build").WithType(cli.TypeMultiString)).
		WithOption(cli.NewOption("exclude", "choose which roots are excluded from the build").WithType(cli.TypeMultiString)).
		WithOption(cli.NewOption("scope", "set the scope for the command")).
		WithAction(_scopeHandler(
			func(req cli.ActionRequest) int {
				var args = req.Args
				var options = req.Opts

				var path string
				var err error

				if len(args) > 0 {
					path = args[0]
				}

				var includeOpts []string
				var excludeOpts []string

				if options["exclude"] != nil {
					excludeOpts = options["exclude"].([]string)
				}

				if options["include"] != nil {
					includeOpts = options["include"].([]string)
				}

				if len(args) > 0 {
					path = args[0]
				}

				includeRoots := dryad.RootIncludeMatcher(includeOpts)
				excludeRoots := dryad.RootExcludeMatcher(excludeOpts)

				err = dryad.GardenBuild(
					dryad.BuildContext{
						RootFingerprints: map[string]string{},
					},
					dryad.GardenBuildRequest{
						BasePath:     path,
						IncludeRoots: includeRoots,
						ExcludeRoots: excludeRoots,
					},
				)
				if err != nil {
					log.Fatal(err)
				}

				return 0
			},
		))

	var gardenPrune = cli.NewCommand("prune", "clear all build artifacts out of the garden not actively linked to a sprout or a root").
		WithAction(func(req cli.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			err = dryad.GardenPrune(
				path,
			)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var gardenPack = cli.NewCommand("pack", "pack the current garden into an archive ").
		WithArg(cli.NewArg("gardenPath", "the path to the garden to pack").AsOptional()).
		WithArg(cli.NewArg("targetPath", "the path (including name) to output the archive to").AsOptional()).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var gardenPath = ""
			var targetPath = ""
			switch len(args) {
			case 0:
				break
			case 1:
				gardenPath = args[0]
			default:
				gardenPath = args[0]
				targetPath = args[1]
			}

			targetPath, err := dryad.GardenPack(gardenPath, targetPath)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(targetPath)
			return 0
		})

	var gardenWipe = cli.NewCommand("wipe", "clear all build artifacts out of the garden").
		WithAction(func(req cli.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			err = dryad.GardenWipe(
				path,
			)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var garden = cli.NewCommand("garden", "commands to work with a dryad garden").
		WithCommand(gardenBuild).
		WithCommand(gardenInit).
		WithCommand(gardenPack).
		WithCommand(gardenPath).
		WithCommand(gardenPrune).
		WithCommand(gardenWipe)

	var rootAdd = cli.NewCommand("add", "add a root as a dependency of the current root").
		WithArg(cli.NewArg("path", "path to the root you want to add as a dependency")).
		WithArg(cli.NewArg("alias", "the alias to add the root under. if not specified, this defaults to the basename of the added root").AsOptional()).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var rootPath, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			var path = args[0]
			var alias = ""
			if len(args) > 1 {
				alias = args[1]
			}

			err = dryad.RootAdd(rootPath, path, alias)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var rootInit = cli.NewCommand("init", "create a new root directory structure in the current dir").
		WithArg(cli.NewArg("path", "the path to init the root at. defaults to current directory").AsOptional()).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var path string = ""

			if len(args) > 0 {
				path = args[0]
			}

			err := dryad.RootInit(path)

			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var rootPath = cli.NewCommand("path", "return the base path of the current root").
		WithArg(cli.NewArg("path", "the path to start searching for a root at. defaults to current directory").AsOptional()).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var path string = ""

			if len(args) > 0 {
				path = args[0]
			}

			path, err := dryad.RootPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)

			return 0
		})

	var rootBuild = cli.NewCommand("build", "build a specified root").
		WithArg(cli.NewArg("path", "path to the root to build").AsOptional()).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var path string

			if len(args) > 0 {
				path = args[0]
			}

			if !filepath.IsAbs(path) {
				wd, err := os.Getwd()
				if err != nil {
					log.Fatal(err)
				}
				path = filepath.Join(wd, path)
			}

			var rootFingerprint string
			rootFingerprint, err := dryad.RootBuild(
				dryad.BuildContext{
					RootFingerprints: map[string]string{},
				},
				path,
			)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(rootFingerprint)

			return 0
		})

	var rootCopy = cli.NewCommand("copy", "make a copy of the specified root at a new location").
		WithArg(cli.NewArg("source", "path to the source root")).
		WithArg(cli.NewArg("destination", "destination path for the root copy")).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var source string = args[0]
			var dest string = args[1]

			err := dryad.RootCopy(source, dest)

			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var root = cli.NewCommand("root", "commands to work with a dryad root").
		WithCommand(rootAdd).
		WithCommand(rootBuild).
		WithCommand(rootCopy).
		WithCommand(rootInit).
		WithCommand(rootPath)

	var rootsList = cli.NewCommand("list", "list all roots that are dependencies for the current root (or roots of the current garden, if the path is not a root)").
		WithArg(cli.NewArg("path", "path to the base root (or garden) to list roots in").AsOptional()).
		WithOption(cli.NewOption("include", "choose which roots are included in the list").WithType(cli.TypeMultiString)).
		WithOption(cli.NewOption("exclude", "choose which roots are excluded from the list").WithType(cli.TypeMultiString)).
		WithOption(cli.NewOption("scope", "set the scope for the command")).
		WithAction(_scopeHandler(
			func(req cli.ActionRequest) int {
				var args = req.Args
				var options = req.Opts

				var path string = ""
				var err error

				if len(args) > 0 {
					path = args[0]
				}

				var gardenPath string
				gardenPath, err = dryad.GardenPath(path)
				if err != nil {
					log.Fatal(err)
				}

				var includeOpts []string
				var excludeOpts []string

				if options["exclude"] != nil {
					excludeOpts = options["exclude"].([]string)
				}

				if options["include"] != nil {
					includeOpts = options["include"].([]string)
				}

				includeRoots := dryad.RootIncludeMatcher(includeOpts)
				excludeRoots := dryad.RootExcludeMatcher(excludeOpts)

				err = dryad.RootsWalk(path, func(path string, info fs.FileInfo) error {

					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(gardenPath, path)
					if err != nil {
						return err
					}

					if includeRoots(relPath) && !excludeRoots(relPath) {
						fmt.Println(path)
					}

					return nil
				})
				if err != nil {
					log.Fatal(err)
				}

				return 0
			},
		))

	var rootsPath = cli.NewCommand("path", "return the path of the roots dir").
		WithAction(func(req cli.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.RootsPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)

			return 0
		})

	var roots = cli.NewCommand("roots", "commands to work with dryad roots").
		WithCommand(rootsList).
		WithCommand(rootsPath)

	var scopeCreate = cli.NewCommand("create", "create a new scope directory for the garden").
		WithArg(cli.NewArg("name", "the name of the new scope")).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var name string = args[0]

			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			scopePath, err := dryad.ScopeCreate(path, name)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(scopePath)

			return 0
		})

	var scopeDelete = cli.NewCommand("delete", "remove an existing scope directory from the garden").
		WithArg(cli.NewArg("name", "the name of the scope to delete")).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var name string = args[0]

			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			err = dryad.ScopeDelete(path, name)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var scopeSettingGet = cli.NewCommand("get", "print the value of a setting in a scope, if it exists").
		WithArg(cli.NewArg("scope", "the name of the scope")).
		WithArg(cli.NewArg("setting", "the name of the setting")).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var scope string = args[0]
			var setting string = args[1]

			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			value, err := dryad.ScopeSettingGet(path, scope, setting)
			if err != nil {
				log.Fatal(err)
			}

			if value != "" {
				fmt.Println(value)
			}

			return 0
		})

	var scopeSettingSet = cli.NewCommand("set", "set the value of a setting in a scope").
		WithArg(cli.NewArg("scope", "the name of the scope")).
		WithArg(cli.NewArg("setting", "the name of the setting")).
		WithArg(cli.NewArg("value", "the new value for the setting")).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var scope string = args[0]
			var setting string = args[1]
			var value string = args[2]

			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			err = dryad.ScopeSettingSet(path, scope, setting, value)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var scopeSettingUnset = cli.NewCommand("unset", "remove a setting from a scope").
		WithArg(cli.NewArg("scope", "the name of the scope")).
		WithArg(cli.NewArg("setting", "the name of the setting")).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var scope string = args[0]
			var setting string = args[1]

			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			err = dryad.ScopeSettingUnset(path, scope, setting)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var scopeSetting = cli.NewCommand("setting", "commands to work with scope settings").
		WithCommand(scopeSettingGet).
		WithCommand(scopeSettingSet).
		WithCommand(scopeSettingUnset)

	var scope = cli.NewCommand("scope", "commands to work with a single scope").
		WithCommand(scopeCreate).
		WithCommand(scopeDelete).
		WithCommand(scopeSetting)

	var scopesDefaultGet = cli.NewCommand("get", "return the name of the default scope, if set").
		WithAction(func(req cli.ActionRequest) int {

			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			scopeName, err := dryad.ScopeGetDefault(path)
			if err != nil {
				log.Fatal(err)
			}

			if scopeName != "" {
				fmt.Println(scopeName)
			}

			return 0
		})

	var scopesDefaultSet = cli.NewCommand("set", "set a scope to be the default").
		WithArg(cli.NewArg("name", "the name of the scope to set as default")).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var name string = args[0]

			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			err = dryad.ScopeSetDefault(path, name)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var scopesDefaultUnset = cli.NewCommand("unset", "remove the default scope setting").
		WithAction(func(req cli.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			err = dryad.ScopeUnsetDefault(path)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})
	var scopesDefault = cli.NewCommand("default", "work with the default scope").
		WithCommand(scopesDefaultGet).
		WithCommand(scopesDefaultSet).
		WithCommand(scopesDefaultUnset)

	var scopesList = cli.NewCommand("list", "list all scopes in the current garden").
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var path string = ""
			var err error

			if len(args) > 0 {
				path = args[0]
			}

			err = dryad.ScopesWalk(path, func(path string, info fs.FileInfo) error {
				fmt.Println(filepath.Base(path))
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var scopesPath = cli.NewCommand("path", "return the path of the scopes dir").
		WithAction(func(req cli.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.ScopesPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)

			return 0
		})

	var scopes = cli.NewCommand("scopes", "commands to work with scopes").
		WithCommand(scopesDefault).
		WithCommand(scopesList).
		WithCommand(scopesPath)

	var scriptRunAction = func(req cli.ActionRequest) int {
		var command = req.Args[0]
		var args = req.Args[1:]
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

		// if the scope is unset, bypass expansion and run the action directly
		if scope == "" || scope == "none" {
			log.Fatal("no scope set, can't find command")
		} else {
			fmt.Println("[info] using scope:", scope)
		}

		var inherit bool
		var env = map[string]string{}

		if options["inherit"] != nil {
			inherit = options["inherit"].(bool)
		} else {
			inherit = true
		}

		// pull environment variables from parent process
		if inherit {
			for _, e := range os.Environ() {
				if i := strings.Index(e, "="); i >= 0 {
					env[e[:i]] = e[i+1:]
				}
			}
		} else {
			// copy a few variables over from parent env for convenience
			env["TERM"] = os.Getenv("TERM")
		}

		err = dryad.ScriptRun(dryad.ScriptRunRequest{
			BasePath: basePath,
			Scope:    scope,
			Setting:  "script-run-" + command,
			Args:     args,
			Env:      env,
		})
		if err != nil {
			log.Fatal(err)
		}

		return 0
	}

	scriptRunAction = _scopeHandler(scriptRunAction)

	var scriptRun = cli.NewCommand("run", "run a script in the current scope").
		WithArg(cli.NewArg("command", "the script name").WithType(cli.TypeString)).
		WithOption(cli.NewOption("scope", "set the scope for the command")).
		WithOption(cli.NewOption("inherit (default true)", "pass all environment variables from the parent environment to the alias to exec").WithType(cli.TypeBool)).
		WithArg(cli.NewArg("-- args", "args to pass to the script").AsOptional()).
		WithAction(scriptRunAction)

	var scriptPathAction = func(req cli.ActionRequest) int {
		var command = req.Args[0]
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

		// if the scope is unset, bypass expansion and run the action directly
		if scope == "" || scope == "none" {
			log.Fatal("no scope set, can't find command")
		} else {
			fmt.Println("[info] using scope:", scope)
		}

		scriptPath, err := dryad.ScriptPath(dryad.ScriptPathRequest{
			BasePath: basePath,
			Scope:    scope,
			Setting:  "script-run-" + command,
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(scriptPath)

		return 0
	}

	scriptPathAction = _scopeHandler(scriptPathAction)

	var scriptPath = cli.NewCommand("path", "print the path to a script").
		WithArg(cli.NewArg("command", "the script name").WithType(cli.TypeString)).
		WithOption(cli.NewOption("scope", "set the scope for the command")).
		WithAction(scriptPathAction)

	var scriptGetAction = func(req cli.ActionRequest) int {
		var command = req.Args[0]
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

		// if the scope is unset, bypass expansion and run the action directly
		if scope == "" || scope == "none" {
			log.Fatal("no scope set, can't find command")
		} else {
			fmt.Println("[info] using scope:", scope)
		}

		script, err := dryad.ScriptGet(dryad.ScriptGetRequest{
			BasePath: basePath,
			Scope:    scope,
			Setting:  "script-run-" + command,
		})
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(script)

		return 0
	}

	scriptGetAction = _scopeHandler(scriptGetAction)

	var scriptGet = cli.NewCommand("get", "print the contents of a script").
		WithArg(cli.NewArg("command", "the script name").WithType(cli.TypeString)).
		WithOption(cli.NewOption("scope", "set the scope for the command")).
		WithAction(scriptGetAction)

	var scriptEditAction = func(req cli.ActionRequest) int {
		var command = req.Args[0]
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

		// if the scope is unset, bypass expansion and run the action directly
		if scope == "" || scope == "none" {
			log.Fatal("no scope set, can't find command")
		} else {
			fmt.Println("[info] using scope:", scope)
		}

		var env = map[string]string{}

		for _, e := range os.Environ() {
			if i := strings.Index(e, "="); i >= 0 {
				env[e[:i]] = e[i+1:]
			}
		}

		if options["editor"] != nil {
			editor := options["editor"].(string)
			env["EDITOR"] = editor
		}

		err = dryad.ScriptEdit(dryad.ScriptEditRequest{
			BasePath: basePath,
			Scope:    scope,
			Setting:  "script-run-" + command,
			Env:      env,
		})
		if err != nil {
			log.Fatal(err)
		}

		return 0
	}

	scriptEditAction = _scopeHandler(scriptEditAction)

	var scriptEdit = cli.NewCommand("edit", "edit a script").
		WithArg(cli.NewArg("command", "the script name").WithType(cli.TypeString)).
		WithOption(cli.NewOption("scope", "set the scope for the command")).
		WithOption(cli.NewOption("editor", "set the editor to use")).
		WithAction(scriptEditAction)

	var script = cli.NewCommand("script", "commands to work with a scoped script").
		WithCommand(scriptEdit).
		WithCommand(scriptGet).
		WithCommand(scriptPath).
		WithCommand(scriptRun)

	var run = cli.NewCommand("run", "alias for `dryad script run`").
		WithArg(cli.NewArg("command", "alias command").WithType(cli.TypeString)).
		WithOption(cli.NewOption("scope", "set the scope for the command")).
		WithOption(cli.NewOption("inherit (default true)", "pass all environment variables from the parent environment to the alias to exec").WithType(cli.TypeBool)).
		WithArg(cli.NewArg("-- args", "args to pass to the command").AsOptional()).
		WithAction(scriptRunAction)

	var scriptsListAction = func(req cli.ActionRequest) int {
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

		err = dryad.ScriptsWalk(dryad.ScriptsWalkRequest{
			BasePath: basePath,
			Scope:    scope,
			OnMatch: func(path string, info fs.FileInfo) error {
				if showPath {
					fmt.Println(path)
				} else {
					var name string = info.Name()
					fmt.Println("dryad script run", strings.TrimPrefix(name, "script-run-"))
				}
				return nil
			},
		})
		if err != nil {
			log.Fatal(err)
		}

		return 0
	}

	scriptsListAction = _scopeHandler(scriptsListAction)

	var scriptsList = cli.NewCommand("list", "list all available scripts in the current scope").
		WithOption(cli.NewOption("scope", "set the scope for the command")).
		WithOption(cli.NewOption("path", "print the path to the scripts instead of the script run command").WithType(cli.TypeBool)).
		WithAction(scriptsListAction)

	var scripts = cli.NewCommand("scripts", "commands to work with scoped scripts").
		WithCommand(scriptsList)

	var secretsFingerprint = cli.NewCommand("fingerprint", "calculate the fingerprint for the secrets in a stem/root").
		WithArg(cli.NewArg("path", "path to the stem base dir")).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var err error
			var path string

			if len(args) > 0 {
				path = args[0]
				path, err = filepath.Abs(path)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				path, err = os.Getwd()
				if err != nil {
					log.Fatal(err)
				}
			}

			// normalize the path to point to the closest secrets
			path, err = dryad.SecretsPath(path)
			if err != nil {
				log.Fatal(err)
			}

			fingerprint, err := dryad.SecretsFingerprint(
				dryad.SecretsFingerprintArgs{
					BasePath: path,
				},
			)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(fingerprint)

			return 0
		})

	var secretsList = cli.NewCommand("list", "list the secret files in a stem/root").
		WithArg(cli.NewArg("path", "path to the stem base dir")).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var err error
			var path string

			if len(args) > 0 {
				path = args[0]
				path, err = filepath.Abs(path)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				path, err = os.Getwd()
				if err != nil {
					log.Fatal(err)
				}
			}

			// normalize the path to point to the closest secrets
			path, err = dryad.SecretsPath(path)
			if err != nil {
				log.Fatal(err)
			}

			err = dryad.SecretsWalk(
				dryad.SecretsWalkArgs{
					BasePath: path,
					OnMatch: func(path string, info fs.FileInfo) error {
						fmt.Println(path)
						return nil
					},
				},
			)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var secretsPath = cli.NewCommand("path", "print the path to the secrets for the current package, if it exists").
		WithArg(cli.NewArg("path", "path to the stem base dir")).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var err error
			var path string

			if len(args) > 0 {
				path = args[0]
				path, err = filepath.Abs(path)
				if err != nil {
					log.Fatal(err)
				}
			} else {
				path, err = os.Getwd()
				if err != nil {
					log.Fatal(err)
				}
			}

			// normalize the path to point to the closest secrets
			path, err = dryad.SecretsPath(path)
			if err != nil {
				log.Fatal(err)
			}

			// check if the secrets folder exists
			exists, err := dryad.SecretsExist(path)
			if err != nil {
				log.Fatal(err)
			}

			if exists {
				fmt.Println(path)
			}

			return 0
		})

	var secrets = cli.NewCommand("secrets", "commands to work with dryad secrets").
		WithCommand(secretsFingerprint).
		WithCommand(secretsList).
		WithCommand(secretsPath)

	var sproutsRun = cli.NewCommand("run", "run each sprout in the current garden").
		WithOption(cli.NewOption("include", "choose which sprouts are included").WithType(cli.TypeMultiString)).
		WithOption(cli.NewOption("exclude", "choose which sprouts are excluded").WithType(cli.TypeMultiString)).
		WithOption(cli.NewOption("context", "name of the execution context. the HOME env var is set to the path for this context")).
		WithOption(cli.NewOption("inherit", "pass all environment variables from the parent environment to the stem").WithType(cli.TypeBool)).
		WithOption(cli.NewOption("confirm", "display the list of sprouts to exec, and ask for confirmation").WithType(cli.TypeBool)).
		WithOption(cli.NewOption("ignore-errors", "continue running even if a sprout returns an error").WithType(cli.TypeBool)).
		WithOption(cli.NewOption("scope", "set the scope for the command")).
		WithArg(cli.NewArg("-- args", "args to pass to each sprout on execution").AsOptional()).
		WithAction(_scopeHandler(
			func(req cli.ActionRequest) int {
				var args = req.Args
				var options = req.Opts

				var path string = ""
				var err error

				if len(args) > 0 {
					path = args[0]
				}

				var gardenPath string
				gardenPath, err = dryad.GardenPath(path)
				if err != nil {
					log.Fatal(err)
				}

				var includeOpts []string
				var excludeOpts []string

				if options["exclude"] != nil {
					excludeOpts = options["exclude"].([]string)
				}

				if options["include"] != nil {
					includeOpts = options["include"].([]string)
				}

				includeSprouts := dryad.RootIncludeMatcher(includeOpts)
				excludeSprouts := dryad.RootExcludeMatcher(excludeOpts)

				var context string
				var inherit bool
				var ignoreErrors bool
				var confirm bool

				if options["context"] != nil {
					context = options["context"].(string)
				}

				if options["inherit"] != nil {
					inherit = options["inherit"].(bool)
				}

				if options["ignore-errors"] != nil {
					ignoreErrors = options["ignore-errors"].(bool)
				}

				if options["confirm"] != nil {
					confirm = options["confirm"].(bool)
				}

				// if confirm is set, we want to print the list
				// of sprouts to run
				if confirm {
					fmt.Println("[warn] dryad sprouts exec will execute these sprouts:")

					err = dryad.SproutsWalk(path, func(path string, info fs.FileInfo) error {

						// calculate the relative path to the root from the base of the garden
						relPath, err := filepath.Rel(gardenPath, path)
						if err != nil {
							return err
						}

						if includeSprouts(relPath) && !excludeSprouts(relPath) {
							fmt.Println("[warn] - " + path)
						}

						return nil
					})
					if err != nil {
						log.Fatal(err)
					}

					fmt.Println("[warn] are you sure? y/n")

					reader := bufio.NewReader(os.Stdin)

					input, err := reader.ReadString('\n')
					if err != nil {
						fmt.Println("[error] error reading input", err)
						return -1
					}

					input = strings.TrimSuffix(input, "\n")

					if input != "y" {
						fmt.Println("[warn] confirmation denied, aborting")
						return 0
					}

				}

				var env = map[string]string{}

				// pull environment variables from parent process
				if inherit {
					for _, e := range os.Environ() {
						if i := strings.Index(e, "="); i >= 0 {
							env[e[:i]] = e[i+1:]
						}
					}
				} else {
					// copy a few variables over from parent env for convenience
					env["TERM"] = os.Getenv("TERM")
				}

				extras := args[0:]

				err = dryad.SproutsWalk(path, func(path string, info fs.FileInfo) error {

					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(gardenPath, path)
					if err != nil {
						return err
					}

					if includeSprouts(relPath) && !excludeSprouts(relPath) {
						fmt.Println("[info] running sprout at", path)

						err := dryad.StemRun(dryad.StemRunRequest{
							StemPath:   path,
							Env:        env,
							Args:       extras,
							JoinStdout: true,
							Context:    context,
						})
						if err != nil {
							if ignoreErrors {
								fmt.Println("[warn] sprout at", path, "threw error", err)
							} else {
								return err
							}
						}

					}

					return nil
				})
				if err != nil {
					log.Fatal(err)
				}

				return 0
			},
		))

	var sproutsList = cli.NewCommand("list", "list all sprouts of the current garden").
		WithOption(cli.NewOption("include", "choose which sprouts are included in the list").WithType(cli.TypeMultiString)).
		WithOption(cli.NewOption("exclude", "choose which sprouts are excluded from the list").WithType(cli.TypeMultiString)).
		WithOption(cli.NewOption("scope", "set the scope for the command")).
		WithAction(_scopeHandler(
			func(req cli.ActionRequest) int {
				var args = req.Args
				var options = req.Opts

				var path string = ""
				var err error

				if len(args) > 0 {
					path = args[0]
				}

				var gardenPath string
				gardenPath, err = dryad.GardenPath(path)
				if err != nil {
					log.Fatal(err)
				}

				var includeOpts []string
				var excludeOpts []string

				if options["exclude"] != nil {
					excludeOpts = options["exclude"].([]string)
				}

				if options["include"] != nil {
					includeOpts = options["include"].([]string)
				}

				includeSprouts := dryad.RootIncludeMatcher(includeOpts)
				excludeSprouts := dryad.RootExcludeMatcher(excludeOpts)

				err = dryad.SproutsWalk(path, func(path string, info fs.FileInfo) error {

					// calculate the relative path to the root from the base of the garden
					relPath, err := filepath.Rel(gardenPath, path)
					if err != nil {
						return err
					}

					if includeSprouts(relPath) && !excludeSprouts(relPath) {
						fmt.Println(path)
					}

					return nil
				})
				if err != nil {
					log.Fatal(err)
				}

				return 0
			},
		))

	var sproutsPath = cli.NewCommand("path", "return the path of the sprouts dir").
		WithAction(func(req cli.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.SproutsPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)

			return 0
		})

	var sprouts = cli.NewCommand("sprouts", "commands to work with dryad sprouts").
		WithCommand(sproutsList).
		WithCommand(sproutsPath).
		WithCommand(sproutsRun)

	var stemRun = cli.NewCommand("run", "execute the main for a stem").
		WithArg(cli.NewArg("path", "path to the stem base dir")).
		WithOption(cli.NewOption("execPath", "path to the executable running `dryad stem run`. used for path setting")).
		WithOption(cli.NewOption("context", "name of the execution context. the HOME env var is set to the path for this context")).
		WithOption(cli.NewOption("inherit", "pass all environment variables from the parent environment to the stem").WithType(cli.TypeBool)).
		WithArg(cli.NewArg("-- args", "args to pass to the stem").AsOptional()).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args
			var options = req.Opts

			var execPath string
			var context string
			var inherit bool

			if options["execPath"] != nil {
				execPath = options["execPath"].(string)
			}

			if options["context"] != nil {
				context = options["context"].(string)
			}

			if options["inherit"] != nil {
				inherit = options["inherit"].(bool)
			}

			var env = map[string]string{}

			// pull
			if inherit {
				for _, e := range os.Environ() {
					if i := strings.Index(e, "="); i >= 0 {
						env[e[:i]] = e[i+1:]
					}
				}
			} else {
				// copy a few variables over from parent env for convenience
				env["TERM"] = os.Getenv("TERM")
			}

			path := args[0]
			extras := args[1:]
			err := dryad.StemRun(dryad.StemRunRequest{
				ExecPath:   execPath,
				StemPath:   path,
				Env:        env,
				Args:       extras,
				JoinStdout: true,
				Context:    context,
			})
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var stemFingerprint = cli.NewCommand("fingerprint", "calculate the fingerprint for a stem dir").
		WithArg(cli.NewArg("path", "path to the stem base dir").AsOptional()).
		WithOption(cli.NewOption("exclude", "a regular expression to exclude files from the fingerprint calculation. the regexp matches against the file path relative to the stem base directory")).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args
			var options = req.Opts

			var err error
			var matchExclude *regexp.Regexp

			if options["exclude"] != "" {
				matchExclude, err = regexp.Compile(options["exclude"].(string))
				if err != nil {
					log.Fatal(err)
				}
			}

			var path string
			if len(args) > 0 {
				path = args[0]
			}
			if path == "" {
				path, err = os.Getwd()
				if err != nil {
					log.Fatal(err)
				}
			}

			var fingerprintString, fingerprintErr = dryad.StemFingerprint(
				dryad.StemFingerprintArgs{
					BasePath:  path,
					MatchDeny: matchExclude,
				},
			)
			if fingerprintErr != nil {
				log.Fatal(fingerprintErr)
			}
			fmt.Println(fingerprintString)

			return 0
		})

	var stemFiles = cli.NewCommand("files", "list the files in a stem").
		// WithArg(cli.NewArg("path", "path to the stem base dir")).
		WithOption(cli.NewOption("exclude", "a regular expression to exclude files from the list. the regexp matches against the file path relative to the stem base directory")).
		WithAction(func(req cli.ActionRequest) int {
			var options = req.Opts

			var err error
			var matchExclude *regexp.Regexp

			if options["exclude"] != "" {
				matchExclude, err = regexp.Compile(options["exclude"].(string))
				if err != nil {
					log.Fatal(err)
				}
			}

			var path string
			path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			err = dryad.StemFiles(
				dryad.StemFilesArgs{
					BasePath:  path,
					MatchDeny: matchExclude,
				},
			)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var stemPack = cli.NewCommand("pack", "pack the stem at the target path into a tar archive").
		WithArg(cli.NewArg("stemPath", "the path to the stem to pack")).
		WithArg(cli.NewArg("targetPath", "the path (including name) to output the archive to").AsOptional()).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var stemPath = args[0]
			var targetPath = ""
			if len(args) > 1 {
				targetPath = args[1]
			}

			targetPath, err := dryad.StemPack(stemPath, targetPath)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(targetPath)
			return 0
		})

	var stemPath = cli.NewCommand("path", "return the base path of the current root").
		// WithArg(cli.NewArg("path", "path to the stem base dir")).
		WithAction(func(req cli.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.StemPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)

			return 0
		})

	var stemUnpack = cli.NewCommand("unpack", "unpack a stem archive at the target path and import it into the current garden").
		WithArg(cli.NewArg("archive", "the path to the archive to unpack")).
		WithAction(func(req cli.ActionRequest) int {
			var args = req.Args

			var stemPath = args[0]

			gardenPath, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}

			targetPath, err := dryad.StemUnpack(gardenPath, stemPath)
			if err != nil {
				log.Fatal(err)
			}

			fmt.Println(targetPath)
			return 0
		})

	var stem = cli.NewCommand("stem", "commands to work with dryad stems").
		WithCommand(stemFingerprint).
		WithCommand(stemFiles).
		WithCommand(stemPack).
		WithCommand(stemPath).
		WithCommand(stemRun).
		WithCommand(stemUnpack)

	var stemsList = cli.NewCommand("list", "list all stems that are dependencies for the current root").
		WithAction(func(req cli.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			err = dryad.StemsWalk(path, func(path string, info fs.FileInfo, err error) error {
				fmt.Println(path)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var stemsPath = cli.NewCommand("path", "return the path of the stems dir").
		WithAction(func(req cli.ActionRequest) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.StemsPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)

			return 0
		})

	var stems = cli.NewCommand("stems", "commands to work with dryad stems").
		WithCommand(stemsList).
		WithCommand(stemsPath)

	var version = cli.NewCommand("version", "print out detailed version info").
		WithAction(func(req cli.ActionRequest) int {
			fmt.Println("version=" + Version)
			fmt.Println("source_fingerprint=" + Fingerprint)
			fmt.Println("arch=" + runtime.GOARCH)
			fmt.Println("os=" + runtime.GOOS)
			return 0
		})

	app = app.
		WithCommand(garden).
		WithCommand(root).
		WithCommand(roots).
		WithCommand(run).
		WithCommand(scope).
		WithCommand(scopes).
		WithCommand(script).
		WithCommand(scripts).
		WithCommand(secrets).
		WithCommand(sprouts).
		WithCommand(stem).
		WithCommand(stems).
		WithCommand(version)

	return app
}

func main() {
	app := _buildCLI()

	// lie to cli about the name of the tool,
	// so that the help always shows the name of the command as
	// `dryad`
	args := os.Args
	args[0] = "dryad"
	os.Exit(app.Run(args, os.Stdout))
}
