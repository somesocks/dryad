package core

type UnsafeShedReference struct {
	BasePath string
	Garden   *SafeGardenReference
}

type SafeShedReference struct {
	BasePath string
	Garden   *SafeGardenReference
}
