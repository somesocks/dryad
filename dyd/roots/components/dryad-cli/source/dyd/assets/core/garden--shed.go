package core

import "dryad/internal/filepath"

func (sg *SafeGardenReference) Shed() *UnsafeShedReference {
	var shedRef = UnsafeShedReference{
		BasePath: filepath.Join(sg.BasePath, "dyd", "shed"),
		Garden:   sg,
	}
	return &shedRef
}

func safeShedReference(garden *SafeGardenReference) *SafeShedReference {
	var shedRef = SafeShedReference{
		BasePath: filepath.Join(garden.BasePath, "dyd", "shed"),
		Garden:   garden,
	}
	return &shedRef
}
