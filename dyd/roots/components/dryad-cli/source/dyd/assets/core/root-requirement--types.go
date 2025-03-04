package core

type UnsafeRootRequirementReference struct {
	BasePath string
	Requirements *SafeRootRequirementsReference
}

type SafeRootRequirementReference struct {
	BasePath string
	Requirements *SafeRootRequirementsReference
}
