package core

type UnsafeHeapSecretReference struct {
	BasePath string
	Secrets *SafeHeapSecretsReference
}

type SafeHeapSecretReference struct {
	BasePath string
	Fingerprint string
	Secrets *SafeHeapSecretsReference
}

