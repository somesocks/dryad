package core

type UnsafeHeapSproutsReference struct {
	BasePath string
	Heap     *SafeHeapReference
}

type SafeHeapSproutsReference struct {
	BasePath string
	Heap     *SafeHeapReference
}
