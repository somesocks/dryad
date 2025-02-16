package core

type UnsafeHeapDerivationReference struct {
	BasePath string
	Derivations *SafeHeapDerivationsReference
}

type SafeHeapDerivationReference struct {
	BasePath string
	Source *UnsafeHeapStemReference
	Result *UnsafeHeapStemReference
	Derivations *SafeHeapDerivationsReference
}

