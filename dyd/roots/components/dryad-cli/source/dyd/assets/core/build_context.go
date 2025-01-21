package core

import (
	"sync"
)

type BuildContext struct {
	Fingerprints map[string]string
	FingerprintsMutex sync.Mutex
}
