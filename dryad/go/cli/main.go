package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"

	dryad "dryad/core"

	cli "dryad/cli"
)

func _buildCLI() cli.App {

	var gardenInit = cli.NewCommand("init", "initialize a garden").
		WithAction(func(args []string, options map[string]string) int {
			path, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			err = dryad.GardenInit(path)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var gardenPath = cli.NewCommand("path", "return the base path for a garden").
		WithAction(func(args []string, options map[string]string) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.GardenPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)

			return 0
		})

	var gardenBuild = cli.NewCommand("build", "build all roots in the garden").
		WithAction(func(args []string, options map[string]string) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			err = dryad.GardenBuild(
				dryad.BuildContext{
					RootFingerprints: map[string]string{},
				},
				path,
			)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var gardenWipe = cli.NewCommand("wipe", "clear all build artifacts out of the garden").
		WithAction(func(args []string, options map[string]string) int {
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
		WithCommand(gardenInit).
		WithCommand(gardenBuild).
		WithCommand(gardenWipe).
		WithCommand(gardenPath)

	var rootAdd = cli.NewCommand("add", "add a root as a dependency of the current root").
		WithArg(cli.NewArg("path", "path to the root you want to add as a dependency")).
		WithArg(cli.NewArg("alias", "the alias to add the root under. if not specified, this defaults to the basename of the added root").AsOptional()).
		WithAction(func(args []string, options map[string]string) int {
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
		WithAction(func(args []string, options map[string]string) int {
			var path string = ""

			if len(args) > 0 {
				path = args[0]
			}

			if path == "" {
				var cwd, err = os.Getwd()
				if err != nil {
					log.Fatal(err)
				}

				path = cwd
			} else if !filepath.IsAbs(path) {
				var cwd, err = os.Getwd()
				if err != nil {
					log.Fatal(err)
				}

				path = filepath.Join(cwd, path)
			}

			dryad.RootInit(path)

			return 0
		})

	var rootPath = cli.NewCommand("path", "return the base path of the current root").
		WithAction(func(args []string, options map[string]string) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.RootPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)

			return 0
		})

	var rootBuild = cli.NewCommand("build", "build the current root").
		WithAction(func(args []string, options map[string]string) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			var rootFingerprint string
			rootFingerprint, err = dryad.RootBuild(
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

	var root = cli.NewCommand("root", "commands to work with a dryad root").
		WithCommand(rootAdd).
		WithCommand(rootInit).
		WithCommand(rootBuild).
		WithCommand(rootPath)

	var rootsList = cli.NewCommand("list", "list all roots that are dependencies for the current root").
		WithAction(func(args []string, options map[string]string) int {
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			err = dryad.RootsWalk(path, func(path string, info fs.FileInfo, err error) error {
				fmt.Println(path)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var rootsPath = cli.NewCommand("path", "return the path of the roots dir").
		WithAction(func(args []string, options map[string]string) int {
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

	var stemExec = cli.NewCommand("exec", "execute the main for a stem").
		WithArg(cli.NewArg("path", "path to the stem base dir")).
		WithArg(cli.NewArg("-- args", "args to pass to the stem").AsOptional()).
		WithAction(func(args []string, options map[string]string) int {
			path := args[0]
			extras := args[1:]
			err := dryad.StemExec(path, nil, extras...)
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var stemFingerprint = cli.NewCommand("fingerprint", "calculate the fingerprint for a stem dir").
		// WithArg(cli.NewArg("path", "path to the stem base dir")).
		WithOption(cli.NewOption("exclude", "a regular expression to exclude files from the fingerprint calculation. the regexp matches against the file path relative to the stem base directory")).
		WithAction(func(args []string, options map[string]string) int {
			var err error
			var matchExclude *regexp.Regexp

			if options["exclude"] != "" {
				matchExclude, err = regexp.Compile(options["exclude"])
				if err != nil {
					log.Fatal(err)
				}
			}

			var path string
			path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
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
		WithAction(func(args []string, options map[string]string) int {
			var err error
			var matchExclude *regexp.Regexp

			if options["exclude"] != "" {
				matchExclude, err = regexp.Compile(options["exclude"])
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

	var stemPath = cli.NewCommand("path", "return the base path of the current root").
		// WithArg(cli.NewArg("path", "path to the stem base dir")).
		WithAction(func(args []string, options map[string]string) int {
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

	var stem = cli.NewCommand("stem", "commands to work with dryad stems").
		WithCommand(stemExec).
		WithCommand(stemFingerprint).
		WithCommand(stemFiles).
		WithCommand(stemPath)

	var stemsList = cli.NewCommand("list", "list all stems that are dependencies for the current root").
		WithAction(func(args []string, options map[string]string) int {
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
		WithAction(func(args []string, options map[string]string) int {
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

	var app = cli.New("dryad package manager").
		WithCommand(garden).
		WithCommand(root).
		WithCommand(roots).
		WithCommand(stem).
		WithCommand(stems)

	return app
}

func main() {
	app := _buildCLI()
	os.Exit(app.Run(os.Args, os.Stdout))
}
