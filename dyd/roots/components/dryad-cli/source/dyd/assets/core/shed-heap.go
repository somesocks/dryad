package core

import (
	"dryad/internal/os"
	"dryad/task"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
)

const (
	shedHeapDepthDefault      = 1
	shedHeapFanoutSegmentSize = 2
)

type shedHeapDepthRequest struct {
	BasePath string
	Key      string
}

type shedHeapDepthKey struct {
	Group    string
	BasePath string
	Key      string
}

func shedHeapPath(basePath string, segments ...string) string {
	dydPath := filepath.Dir(filepath.Dir(basePath))
	parts := append([]string{dydPath, "shed", "heap"}, segments...)
	return filepath.Join(parts...)
}

func shedHeapFilesDepthPath(basePath string) string {
	return shedHeapPath(basePath, "files", "depth")
}

func shedHeapSecretsDepthPath(basePath string) string {
	return shedHeapPath(basePath, "secrets", "depth")
}

func shedHeapStemsDepthPath(basePath string) string {
	return shedHeapPath(basePath, "stems", "depth")
}

func shedHeapSproutsDepthPath(basePath string) string {
	return shedHeapPath(basePath, "sprouts", "depth")
}

func shedHeapDerivationsRootsDepthPath(basePath string) string {
	return shedHeapPath(basePath, "derivations", "roots", "depth")
}

func shedHeapDepthRead(ctx *task.ExecutionContext, req shedHeapDepthRequest) (error, int) {
	bytes, err := os.ReadFile(req.BasePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, shedHeapDepthDefault
		}
		return err, 0
	}

	raw := strings.TrimSpace(string(bytes))
	if raw == "" {
		return fmt.Errorf("invalid shed heap depth in %q", req.BasePath), 0
	}

	depth, err := strconv.Atoi(raw)
	if err != nil {
		return fmt.Errorf("invalid shed heap depth in %q: %w", req.BasePath, err), 0
	}
	if depth < 0 {
		return fmt.Errorf("invalid shed heap depth in %q: must be non-negative", req.BasePath), 0
	}

	return nil, depth
}

var memoShedHeapDepthRead = task.Memoize(
	shedHeapDepthRead,
	func(ctx *task.ExecutionContext, req shedHeapDepthRequest) (error, any) {
		return nil, shedHeapDepthKey{
			Group:    "ShedHeapDepthRead",
			BasePath: req.BasePath,
			Key:      req.Key,
		}
	},
)

func shedHeapFilesDepth(ctx *task.ExecutionContext, basePath string) (error, int) {
	return memoShedHeapDepthRead(
		ctx,
		shedHeapDepthRequest{
			BasePath: shedHeapFilesDepthPath(basePath),
			Key:      "files",
		},
	)
}

func shedHeapSecretsDepth(ctx *task.ExecutionContext, basePath string) (error, int) {
	return memoShedHeapDepthRead(
		ctx,
		shedHeapDepthRequest{
			BasePath: shedHeapSecretsDepthPath(basePath),
			Key:      "secrets",
		},
	)
}

func shedHeapStemsDepth(ctx *task.ExecutionContext, basePath string) (error, int) {
	return memoShedHeapDepthRead(
		ctx,
		shedHeapDepthRequest{
			BasePath: shedHeapStemsDepthPath(basePath),
			Key:      "stems",
		},
	)
}

func shedHeapSproutsDepth(ctx *task.ExecutionContext, basePath string) (error, int) {
	return memoShedHeapDepthRead(
		ctx,
		shedHeapDepthRequest{
			BasePath: shedHeapSproutsDepthPath(basePath),
			Key:      "sprouts",
		},
	)
}

func shedHeapDerivationsRootsDepth(ctx *task.ExecutionContext, basePath string) (error, int) {
	return memoShedHeapDepthRead(
		ctx,
		shedHeapDepthRequest{
			BasePath: shedHeapDerivationsRootsDepthPath(basePath),
			Key:      "derivations/roots",
		},
	)
}

func heapFingerprintPath(basePath string, fingerprint string, depth int) (error, string) {
	err, version, encoded := fingerprintParse(fingerprint)
	if err != nil {
		return err, ""
	}
	if depth < 0 {
		return fmt.Errorf("invalid shed heap depth: %d", depth), ""
	}
	if depth > 0 && depth*shedHeapFanoutSegmentSize >= len(encoded) {
		return fmt.Errorf("invalid shed heap depth %d for fingerprint %q", depth, fingerprint), ""
	}

	parts := []string{heapVersionDir(basePath, version)}
	remaining := encoded
	for i := 0; i < depth; i++ {
		parts = append(parts, remaining[:shedHeapFanoutSegmentSize])
		remaining = remaining[shedHeapFanoutSegmentSize:]
	}
	parts = append(parts, remaining)

	return nil, filepath.Join(parts...)
}
