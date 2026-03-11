package core

import (
	"dryad/internal/filepath"
	"dryad/internal/os"
	"dryad/task"
	"fmt"
)

type RootBuildProvenance struct {
	ResultFingerprint string
	Sources           map[string]*RootBuildResult
	Results           map[string]map[string]struct{}
}

type rootBuildProvenanceCollectRequest struct {
	Root *RootBuildResult
}

func rootBuildProvenanceCollect(req rootBuildProvenanceCollectRequest) (error, *RootBuildProvenance) {
	if req.Root == nil {
		return fmt.Errorf("missing root build result for provenance collection"), nil
	}

	provenance := &RootBuildProvenance{
		ResultFingerprint: req.Root.ResultFingerprint,
		Sources:           map[string]*RootBuildResult{},
		Results:           map[string]map[string]struct{}{},
	}

	var collect func(*RootBuildResult) error
	collect = func(node *RootBuildResult) error {
		if node == nil {
			return nil
		}

		existing, exists := provenance.Sources[node.SourceFingerprint]
		if exists {
			if existing.ResultFingerprint != node.ResultFingerprint {
				return fmt.Errorf(
					"conflicting provenance node for source fingerprint %s",
					node.SourceFingerprint,
				)
			}
		} else {
			provenance.Sources[node.SourceFingerprint] = node
		}

		resultSources, exists := provenance.Results[node.ResultFingerprint]
		if !exists {
			resultSources = map[string]struct{}{}
			provenance.Results[node.ResultFingerprint] = resultSources
		}
		resultSources[node.SourceFingerprint] = struct{}{}

		for _, dependency := range node.Dependencies {
			err := collect(dependency)
			if err != nil {
				return err
			}
		}
		return nil
	}

	err := collect(req.Root)
	if err != nil {
		return err, nil
	}

	return nil, provenance
}

type rootBuildProvenanceStemRequest struct {
	Garden      *SafeGardenReference
	BuildResult *RootBuildResult
}

func rootBuildProvenanceStem(ctx *task.ExecutionContext, req rootBuildProvenanceStemRequest) (error, *SafeHeapStemReference) {
	if req.Garden == nil {
		return fmt.Errorf("missing garden for provenance stem"), nil
	}
	if req.BuildResult == nil {
		return fmt.Errorf("missing build result for provenance stem"), nil
	}

	err, heap := req.Garden.Heap().Resolve(ctx)
	if err != nil {
		return err, nil
	}

	err, heapStems := heap.Stems().Resolve(ctx)
	if err != nil {
		return err, nil
	}

	provenanceStemPath, err := os.MkdirTemp("", "dryad-*")
	if err != nil {
		return err, nil
	}
	defer os.RemoveAll(provenanceStemPath)

	err = StemInit(provenanceStemPath)
	if err != nil {
		return err, nil
	}

	traitsPath := filepath.Join(provenanceStemPath, "dyd", "traits")
	err = os.WriteFile(filepath.Join(traitsPath, "kind"), []byte("provenance"), 0o511)
	if err != nil {
		return err, nil
	}
	err = os.WriteFile(filepath.Join(traitsPath, "result-fingerprint"), []byte(req.BuildResult.ResultFingerprint), 0o511)
	if err != nil {
		return err, nil
	}

	err, provenance := rootBuildProvenanceCollect(rootBuildProvenanceCollectRequest{
		Root: req.BuildResult,
	})
	if err != nil {
		return err, nil
	}

	dependenciesPath := filepath.Join(provenanceStemPath, "dyd", "dependencies")
	assetsResultsPath := filepath.Join(provenanceStemPath, "dyd", "assets", "results")
	heapStemsPath := filepath.Join(req.Garden.BasePath, "dyd", "heap", "stems")

	for sourceFingerprint := range provenance.Sources {
		err, dependencySourceStemPath := heapStemsFingerprintPath(
			ctx,
			req.Garden,
			heapStemsPath,
			sourceFingerprint,
		)
		if err != nil {
			return err, nil
		}

		err = os.Symlink(
			dependencySourceStemPath,
			filepath.Join(dependenciesPath, sourceFingerprint),
		)
		if err != nil {
			return err, nil
		}
	}

	for resultFingerprint, resultSources := range provenance.Results {
		resultPath := filepath.Join(assetsResultsPath, resultFingerprint)
		err = os.MkdirAll(resultPath, os.ModePerm)
		if err != nil {
			return err, nil
		}

		for sourceFingerprint := range resultSources {
			err = os.WriteFile(
				filepath.Join(resultPath, sourceFingerprint),
				[]byte{},
				0o511,
			)
			if err != nil {
				return err, nil
			}
		}
	}

	err = rootBuild_requirementsPrepare(provenanceStemPath)
	if err != nil {
		return err, nil
	}

	err, _ = stemFinalize(ctx, provenanceStemPath)
	if err != nil {
		return err, nil
	}

	return heapStems.AddStem(
		ctx,
		HeapAddStemRequest{
			StemPath: provenanceStemPath,
		},
	)
}
