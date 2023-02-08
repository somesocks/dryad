package main

import (
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

	var gardenInit = cli.NewCommand("init", "initialize a garden").
		WithArg(cli.NewArg("path", "the target path at which to initialize the garden").AsOptional()).
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
			fmt.Println("option include", options["include"])

			var path string
			var err error

			if len(args) > 0 {
				path = args[0]
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
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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

	var rootsList = cli.NewCommand("list", "list all roots that are dependencies for the current root (or roots of the current garden, if the path is not a root)").
		WithArg(cli.NewArg("path", "path to the base root (or garden) to list roots in").AsOptional()).
		WithAction(func(args []string, options map[string]interface{}) int {
			var path string = ""
			var err error

			if len(args) > 0 {
				path = args[0]
			}

			err = dryad.RootsWalk(path, func(path string, info fs.FileInfo) error {
				fmt.Println(path)
				return nil
			})
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var rootsPath = cli.NewCommand("path", "return the path of the roots dir").
		WithAction(func(args []string, options map[string]interface{}) int {
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

	var secretsFingerprint = cli.NewCommand("fingerprint", "calculate the fingerprint for the secrets in a stem/root").
		WithArg(cli.NewArg("path", "path to the stem base dir")).
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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

	var stemExec = cli.NewCommand("exec", "execute the main for a stem").
		WithArg(cli.NewArg("path", "path to the stem base dir")).
		WithOption(cli.NewOption("execPath", "path to the executable running `dryad stem exec`. used for path setting")).
		WithOption(cli.NewOption("context", "name of the execution context. the HOME env var is set to the path for this context")).
		WithOption(cli.NewOption("inherit", "pass all environment variables from the parent environment to the stem").WithType(cli.TypeBool)).
		WithArg(cli.NewArg("-- args", "args to pass to the stem").AsOptional()).
		WithAction(func(args []string, options map[string]interface{}) int {
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
				ExecPath:   options["execPath"].(string),
				StemPath:   path,
				Env:        env,
				Args:       extras,
				JoinStdout: true,
				Context:    options["context"].(string),
			})
			if err != nil {
				log.Fatal(err)
			}

			return 0
		})

	var stemFingerprint = cli.NewCommand("fingerprint", "calculate the fingerprint for a stem dir").
		WithArg(cli.NewArg("path", "path to the stem base dir").AsOptional()).
		WithOption(cli.NewOption("exclude", "a regular expression to exclude files from the fingerprint calculation. the regexp matches against the file path relative to the stem base directory")).
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
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
		WithAction(func(args []string, options map[string]interface{}) int {
			fmt.Println("version=" + Version)
			fmt.Println("source_fingerprint=" + Fingerprint)
			fmt.Println("arch=" + runtime.GOARCH)
			fmt.Println("os=" + runtime.GOOS)
			return 0
		})

	var app = cli.New("dryad package manager " + Version).
		WithCommand(garden).
		WithCommand(root).
		WithCommand(roots).
		WithCommand(secrets).
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
