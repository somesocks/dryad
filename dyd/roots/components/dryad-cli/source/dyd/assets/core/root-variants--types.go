package core

type UnsafeRootVariantsReference struct {
	BasePath string
	Root     *SafeRootReference
}

type SafeRootVariantsReference struct {
	BasePath string
	Root     *SafeRootReference
}
