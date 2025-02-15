
package core

type UnsafeGardenReference struct {
	BasePath string
}

type SafeGardenReference struct {
	BasePath string
}

func Garden(basePath string) (*UnsafeGardenReference) {
	var ref = UnsafeGardenReference{
		BasePath: basePath,
	}

	return &ref
}