package cli

import (
	dryad "dryad/core"
	"errors"
	"strings"
)

type parsedRootRef struct {
	Path        string
	Selector    dryad.VariantDescriptor
	HasSelector bool
}

func parseRootRef(raw string) (error, parsedRootRef) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, parsedRootRef{
			Path: "",
		}
	}

	if idx := strings.Index(raw, "?"); idx > -1 {
		if idx == 0 {
			return errors.New("missing root ref path"), parsedRootRef{}
		}

		err, variantContext := dryad.RootVariantContextFromURL(raw[idx:])
		if err != nil {
			return err, parsedRootRef{}
		}

		return nil, parsedRootRef{
			Path:        raw[:idx],
			Selector:    variantContext.Descriptor,
			HasSelector: true,
		}
	}

	if idx := strings.LastIndex(raw, dryad.RootRequirementSelectorSeparator); idx > -1 {
		if idx == 0 {
			return errors.New("missing root ref path"), parsedRootRef{}
		}

		selectorRaw := raw[idx+len(dryad.RootRequirementSelectorSeparator):]
		if selectorRaw == "" {
			return errors.New("malformed variant descriptor"), parsedRootRef{}
		}

		err, variantContext := dryad.RootVariantContextFromFilesystem(selectorRaw)
		if err != nil {
			return err, parsedRootRef{}
		}

		return nil, parsedRootRef{
			Path:        raw[:idx],
			Selector:    variantContext.Descriptor,
			HasSelector: true,
		}
	}

	return nil, parsedRootRef{
		Path: raw,
	}
}
