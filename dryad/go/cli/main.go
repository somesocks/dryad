package main

import (
	"fmt"
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
	default:
		log.Fatal("unrecognized command " + command)
	}
}
