package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"

	dryad "dryad-go/core"
)

func main() {
	var arg1 string
	var arg2 string
	var command string

	switch len(os.Args) {
	case 0:
	case 1:
		arg1 = ""
		arg2 = ""
	case 2:
		arg1 = os.Args[1]
		arg2 = ""
	default:
		arg1 = os.Args[1]
		arg2 = os.Args[2]
	}

	command = arg1 + "::" + arg2

	switch command {
	case "::":
		{
			fmt.Print(
				"\n",
				"dryad commands follow a 'dryad <RESOURCE> <ACTION>' pattern\n\n",
				"resources:\n",
				"  garden\n",
				"  heap\n",
				"  root\n",
				"  roots\n",
				"  stem\n",
				"  stems\n",
				"\n",
				"to see actions for a resource, run 'dryad <RESOUCE>'\n",
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
			var path, err = os.Getwd()
			if err != nil {
				log.Fatal(err)
			}
			dryad.GardenInit(path)
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
		fmt.Println("COMMAND garden build")
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
