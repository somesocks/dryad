

package core

type UnsafeRootReference struct {
	BasePath string
	Garden *SafeGardenReference
}

type SafeRootReference struct {
	BasePath string
	Garden *SafeGardenReference
}

