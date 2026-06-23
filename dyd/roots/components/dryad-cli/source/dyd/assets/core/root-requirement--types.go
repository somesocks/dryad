package core

import "io/fs"

type UnsafeRootRequirementReference struct {
	BasePath     string
	Requirements *SafeRootRequirementsReference
}

type SafeRootRequirementReference struct {
	BasePath     string
	fileInfo     fs.FileInfo
	Requirements *SafeRootRequirementsReference
}
