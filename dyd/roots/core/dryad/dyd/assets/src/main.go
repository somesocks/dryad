package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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

	var gardenPrune = cli.NewCommand("prune", "clear all build artifacts out of the garden not actively linked to a sprout or a root").
		WithAction(func(args []string, options map[string]string) int {
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
		WithAction(func(args []string, options map[string]string) int {
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
		WithCommand(gardenBuild).
		WithCommand(gardenInit).
		WithCommand(gardenPack).
		WithCommand(gardenPath).
		WithCommand(gardenPrune).
		WithCommand(gardenWipe)

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

	var rootBuild = cli.NewCommand("build", "build a specified root").
		WithArg(cli.NewArg("path", "path to the root to build")).
		WithAction(func(args []string, options map[string]string) int {
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

	var root = cli.NewCommand("root", "commands to work with a dryad root").
		WithCommand(rootAdd).
		WithCommand(rootBuild).
		WithCommand(rootInit).
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
		WithOption(cli.NewOption("execPath", "path to the executable running `dryad stem exec`. used for path setting")).
		WithOption(cli.NewOption("context", "name of the execution context. the HOME env var is set to the path for this context")).
		WithOption(cli.NewOption("inherit", "pass all environment variables from the parent environment to the stem").WithType(cli.TypeBool)).
		WithArg(cli.NewArg("-- args", "args to pass to the stem").AsOptional()).
		WithAction(func(args []string, options map[string]string) int {
			var env = map[string]string{}

			// pull
			if options["inherit"] == "true" {
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
			err := dryad.StemExec(dryad.StemExecRequest{
				ExecPath:   options["execPath"],
				StemPath:   path,
				Env:        env,
				Args:       extras,
				JoinStdout: true,
				Context:    options["context"],
			})
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var stemFingerprint = cli.NewCommand("fingerprint", "calculate the fingerprint for a stem dir").
		WithArg(cli.NewArg("path", "path to the stem base dir").AsOptional()).
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

	var stemPack = cli.NewCommand("pack", "pack the stem at the target path into a tar archive").
		WithArg(cli.NewArg("stemPath", "the path to the stem to pack")).
		WithArg(cli.NewArg("targetPath", "the path (including name) to output the archive to").AsOptional()).
		WithAction(func(args []string, options map[string]string) int {
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

	var stemUnpack = cli.NewCommand("unpack", "unpack a stem archive at the target path and import it into the current garden").
		WithArg(cli.NewArg("archive", "the path to the archive to unpack")).
		WithAction(func(args []string, options map[string]string) int {
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
		WithCommand(stemExec).
		WithCommand(stemFingerprint).
		WithCommand(stemFiles).
		WithCommand(stemPack).
		WithCommand(stemPath).
		WithCommand(stemUnpack)

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