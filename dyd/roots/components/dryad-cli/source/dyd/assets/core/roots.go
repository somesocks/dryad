package core

type UnsafeRootsReference struct {
	BasePath string
	Garden *SafeGardenReference
}

type SafeRootsReference struct {
	BasePath string
	Garden *SafeGardenReference
}

