package core

import (
	"fmt"
	"os"
	"path/filepath"

	tasks "dryad/tasks"
)

type GardenCreateRequest struct {
	BasePath string
}

func gardenPrepareRequest(req GardenCreateRequest) (error, GardenCreateRequest) {
	path, err := filepath.Abs(req.BasePath)
	if err != nil {
		return err, req
	}
	req.BasePath = path
	return nil, req
}

func gardenCreateBase(req GardenCreateRequest) (error, GardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeap(req GardenCreateRequest) (error, GardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapFiles(req GardenCreateRequest) (error, GardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "files")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapStems(req GardenCreateRequest) (error, GardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "stems")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapDerivations(req GardenCreateRequest) (error, GardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "derivations")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapContexts(req GardenCreateRequest) (error, GardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "contexts")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateHeapSecrets(req GardenCreateRequest) (error, GardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "heap", "secrets")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateShed(req GardenCreateRequest) (error, GardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "shed")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateShedScopes(req GardenCreateRequest) (error, GardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "shed", "scopes")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateRoots(req GardenCreateRequest) (error, GardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "roots")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateSprouts(req GardenCreateRequest) (error, GardenCreateRequest) {
	path := filepath.Join(req.BasePath, "dyd", "sprouts")
	err := os.MkdirAll(path, os.ModePerm)
	return err, req
}

func gardenCreateTypeFile(req GardenCreateRequest) (error, GardenCreateRequest) {
	// write out type file
	typePath := filepath.Join(req.BasePath, "dyd", "type")

	typeFile, err := os.Create(typePath)
	if err != nil {
		return err, req
	}
	defer typeFile.Close()

	_, err = fmt.Fprint(typeFile, "garden")
	if err != nil {
		return err, req
	}

	return nil, req
}

var GardenCreate = tasks.Series4(
	gardenPrepareRequest,
	gardenCreateBase,
	tasks.Parallel5(
		tasks.Series6(
			gardenCreateHeap,
			gardenCreateHeapFiles,
			gardenCreateHeapStems,
			gardenCreateHeapDerivations,
			gardenCreateHeapContexts,
			gardenCreateHeapSecrets,	
		),
		tasks.Series2(
			gardenCreateShed,
			gardenCreateShedScopes,
		),
		gardenCreateRoots,
		gardenCreateSprouts,
		gardenCreateTypeFile,	
	),
	func (res tasks.Tuple5[GardenCreateRequest, GardenCreateRequest, GardenCreateRequest, GardenCreateRequest, GardenCreateRequest]) (error, GardenCreateRequest) {
		return nil, res.A 
	},
)
