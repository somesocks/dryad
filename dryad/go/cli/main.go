package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"

	dryad "dryad/core"
)

func main() {
	var arg1 string
	var arg2 string
	var arg3 string
	var arg4 string
	var command string

	switch len(os.Args) {
	case 0:
	case 1:
		arg1 = ""
		arg2 = ""
		arg3 = ""
		arg4 = ""
	case 2:
		arg1 = os.Args[1]
		arg2 = ""
		arg3 = ""
		arg4 = ""
	case 3:
		arg1 = os.Args[1]
		arg2 = os.Args[2]
		arg3 = ""
		arg4 = ""
	case 4:
		arg1 = os.Args[1]
		arg2 = os.Args[2]
		arg3 = os.Args[3]
		arg4 = ""
	default:
		arg1 = os.Args[1]
		arg2 = os.Args[2]
		arg3 = os.Args[3]
		arg4 = os.Args[4]
	}

	command = arg1 + "::" + arg2

	switch command {
	case "::":
		{
			fmt.Print(
				"\n",
				"dryad commands follow a 'dryad <resource> <action>' pattern\n\n",
				"resources:\n",
				"  garden\n",
				"  heap\n",
				"  root\n",
				"  roots\n",
				"  stem\n",
				"  stems\n",
				"\n",
				"to see actions for a resource, run 'dryad <resource>'\n",
				"\n",
			)
		}
	case "garden::":
		{
			fmt.Print(
				"\n",
				"dryad garden commands:\n",
				"\n",
				"  dryad garden init\n",
				"    initialize a garden in the current directory\n",
				"\n",
				"  dryad garden path\n",
				"    return the path of the parent garden to this directory\n",
				"\n",
				"  dryad garden build\n",
				"    build all roots in the garden\n",
				"\n",
			)
		}
	case "garden::init":
		{
			path, err := os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			err = dryad.GardenInit(path)
			if err != nil {
				log.Fatal(err)
			}
		}
	case "garden::path":
		{
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.GardenPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)
		}
	case "garden::build":
		{
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			err = dryad.GardenBuild(
				dryad.BuildContext{
					map[string]string{},
				},
				path,
			)
			if err != nil {
				log.Fatal(err)
			}
		}
	case "heap::path":
		{
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.HeapPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)
		}
	case "root::":
		{
			fmt.Print(
				"\n",
				"dryad root commands:\n",
				"\n",
				"  dryad root init\n",
				"    initialize a root in the current directory\n",
				"\n",
				"  dryad root path\n",
				"    return the path of the parent root to this directory\n",
				"\n",
				"  dryad root add <path> <alias?>\n",
				"    add a root as a dependency to this root\n",
				"    <path> - the path to the root to add as a dependency\n",
				"    <alias?> - an optional alias for the dependency. if not specified, the basename to the dependency root folder is used\n",
				"\n",
			)
		}
	case "root::add":
		{
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			err = dryad.RootAdd(path, arg3, arg4)
			if err != nil {
				log.Fatal(err)
			}
		}
	case "root::init":
		{
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			dryad.RootInit(path)
		}
	case "root::path":
		{
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.RootPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)
		}
	case "root::pack":
		{
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.StemPack(path, "")
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)
		}
	case "root::build":
		{
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			var rootFingerprint string
			rootFingerprint, err = dryad.RootBuild(
				dryad.BuildContext{
					map[string]string{},
				},
				path,
			)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(rootFingerprint)
		}
	case "roots::list":
		{
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
		}
	case "roots::path":
		{
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.RootsPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)
		}
	case "stems::list":
		{
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
		}
	case "stems::path":
		{
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.StemsPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)
		}
	case "stem::exec":
		{
			path := arg3
			args := os.Args[4:]
			err := dryad.StemExec(path, args...)
			if err != nil {
				log.Fatal(err)
			}
		}
	case "stem::fingerprint":
		{
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			var fingerprintString, fingerprintErr = dryad.StemFingerprint(path)
			if fingerprintErr != nil {
				log.Fatal(fingerprintErr)
			}
			fmt.Println(fingerprintString)
		}
	case "stem::files":
		{
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			err = dryad.StemFiles(path)
			if err != nil {
				log.Fatal(err)
			}
		}
	case "stem::path":
		{
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			path, err = dryad.StemPath(path)
			if err != nil {
				log.Fatal(err)
			}
			fmt.Println(path)
		}
	default:
		log.Fatal("unrecognized command " + command)
	}
}
