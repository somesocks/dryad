package core

import "fmt"

// RootVariantContext carries a concrete variant descriptor for a root build.
// The descriptor must be in filesystem form semantics (dimension=option,...).
type RootVariantContext struct {
	Descriptor VariantDescriptor
}

func RootVariantContextFromFilesystem(raw string) (error, RootVariantContext) {
	err, normalized := variantDescriptorNormalizeFilesystem(raw)
	if err != nil {
		return err, RootVariantContext{}
	}

	err, descriptor := variantDescriptorParseFilesystem(normalized)
	if err != nil {
		return err, RootVariantContext{}
	}

	return nil, RootVariantContext{Descriptor: descriptor}
}

func (context RootVariantContext) Filesystem() (error, string) {
	return variantDescriptorEncodeFilesystem(context.Descriptor)
}

func (context RootVariantContext) MustFilesystem() string {
	err, encoded := context.Filesystem()
	if err != nil {
		panic(fmt.Sprintf("invalid root variant context: %v", err))
	}
	return encoded
}
