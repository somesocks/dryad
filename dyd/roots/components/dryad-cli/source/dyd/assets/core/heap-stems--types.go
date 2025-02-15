package core

type UnsafeHeapStemsReference struct {
	BasePath string
	Heap *SafeHeapReference
}

type SafeHeapStemsReference struct {
	BasePath string
	Heap *SafeHeapReference
}

