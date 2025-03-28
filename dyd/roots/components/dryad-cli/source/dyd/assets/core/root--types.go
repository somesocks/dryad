

package core

type UnsafeRootReference struct {
	BasePath string
	Roots *SafeRootsReference
}

type SafeRootReference struct {
	BasePath string
	Roots *SafeRootsReference
}

