package core

type UnsafeHeapReference struct {
	BasePath string
	Garden *SafeGardenReference
}

type SafeHeapReference struct {
	BasePath string
	Garden *SafeGardenReference
}

