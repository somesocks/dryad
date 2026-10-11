package core

import "context"

type RootsWalkVariantsRequest struct {
	ShouldMatch func(context.Context, *SafeRootVariantReference) (error, bool)
	OnMatch     func(context.Context, *SafeRootVariantReference) (error, any)
}

func (roots *SafeRootsReference) WalkVariants(ctx context.Context, req RootsWalkVariantsRequest) error {
	if req.ShouldMatch == nil {
		req.ShouldMatch = func(context.Context, *SafeRootVariantReference) (error, bool) {
			return nil, true
		}
	}

	return roots.Walk(
		ctx,
		RootsWalkRequest{
			OnMatch: func(ctx context.Context, root *SafeRootReference) (error, any) {
				err, variants := root.ResolveBuildVariantReferences(
					ctx,
					RootResolveBuildVariantsRequest{},
				)
				if err != nil {
					return err, nil
				}

				for _, variant := range variants {
					err, shouldMatch := req.ShouldMatch(ctx, variant)
					if err != nil {
						return err, nil
					}
					if !shouldMatch {
						continue
					}

					err, _ = req.OnMatch(ctx, variant)
					if err != nil {
						return err, nil
					}
				}

				return nil, nil
			},
		},
	)
}
