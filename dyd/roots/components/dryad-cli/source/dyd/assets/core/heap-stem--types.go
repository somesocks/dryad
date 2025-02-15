package core

type UnsafeHeapStemReference struct {
	BasePath string
	Stems *SafeHeapStemsReference
}

type SafeHeapStemReference struct {
	BasePath string
	Stems *SafeHeapStemsReference
}

