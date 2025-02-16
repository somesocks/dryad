package core

type UnsafeHeapFilesReference struct {
	BasePath string
	Heap *SafeHeapReference
}

type SafeHeapFilesReference struct {
	BasePath string
	Heap *SafeHeapReference
}

