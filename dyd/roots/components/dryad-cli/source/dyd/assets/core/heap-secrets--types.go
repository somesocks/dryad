package core

type UnsafeHeapSecretsReference struct {
	BasePath string
	Heap *SafeHeapReference
}

type SafeHeapSecretsReference struct {
	BasePath string
	Heap *SafeHeapReference
}

