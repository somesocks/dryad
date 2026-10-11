package core

import (
	"context"
	dydfs "dryad/filesystem"
	"dryad/task"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"
)

func TestOperationsRefreshIgnoreReaders(t *testing.T) {
	const hiddenFile = "dyd/assets/ignored/child.txt"
	operations := map[string]func(*testing.T, context.Context, string) bool{
		"walk": func(t *testing.T, ctx context.Context, base string) bool {
			var seen atomic.Bool
			err, _ := StemWalk(ctx, StemWalkRequest{
				BasePath: base,
				OnMatch: func(_ context.Context, node dydfs.Walk6Node) (error, any) {
					if node.VPath == filepath.Join(base, hiddenFile) {
						seen.Store(true)
					}
					return nil, nil
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			return seen.Load()
		},
		"collect development state": func(t *testing.T, ctx context.Context, base string) bool {
			states, err := rootDevelop_collectState(ctx, rootDevelopCollectSpec{BasePath: base, ApplyIgnore: true})
			if err != nil {
				t.Fatal(err)
			}
			_, seen := states[filepath.FromSlash(hiddenFile)]
			return seen
		},
		"development copy": func(t *testing.T, ctx context.Context, base string) bool {
			dest := t.TempDir()
			if err := rootDevelop_copyDir(ctx, base, dest, rootDevelopCopyOptions{ApplyIgnore: true}); err != nil {
				t.Fatal(err)
			}
			_, err := os.Stat(filepath.Join(dest, hiddenFile))
			if err != nil && !os.IsNotExist(err) {
				t.Fatal(err)
			}
			return err == nil
		},
	}
	for name, operation := range operations {
		t.Run(name, func(t *testing.T) {
			base := t.TempDir()
			writeFileForTest(t, filepath.Join(base, hiddenFile), "content")
			writeFileForTest(t, filepath.Join(base, ".dyd-ignore"), "dyd/assets/ignored\n")
			ctx := task.NewContext(context.Background(), 3)
			if operation(t, ctx, base) {
				t.Fatal("ignored file was included")
			}
			writeFileForTest(t, filepath.Join(base, ".dyd-ignore"), "")
			if !operation(t, ctx, base) {
				t.Fatal("second operation reused stale ignore data")
			}
		})
	}
}

func TestFingerprintRefreshesIgnoreReader(t *testing.T) {
	base := t.TempDir()
	writeFileForTest(t, filepath.Join(base, "dyd/assets/ignored/child.txt"), "content")
	writeFileForTest(t, filepath.Join(base, "dyd/assets/visible.txt"), "visible")
	writeFileForTest(t, filepath.Join(base, ".dyd-ignore"), "dyd/assets/ignored\n")
	ctx := task.NewContext(context.Background(), 3)
	err, before := StemFingerprint(ctx, StemFingerprintRequest{BasePath: base})
	if err != nil {
		t.Fatal(err)
	}
	writeFileForTest(t, filepath.Join(base, ".dyd-ignore"), "")
	err, after := StemFingerprint(ctx, StemFingerprintRequest{BasePath: base})
	if err != nil {
		t.Fatal(err)
	}
	if before == after {
		t.Fatal("fingerprint reused stale ignore data")
	}
}

func TestRootDiscoveryRefreshesMemoizer(t *testing.T) {
	base := t.TempDir()
	rootPath := filepath.Join(base, "package")
	if err := os.MkdirAll(rootPath, 0o755); err != nil {
		t.Fatal(err)
	}
	roots := &SafeRootsReference{BasePath: base}
	ctx := task.NewContext(context.Background(), 1)
	var calls int
	callbackWork := task.Memoize(
		func(context.Context, int) (error, int) { calls++; return nil, calls },
		func(_ context.Context, req int) (error, any) { return nil, req },
	)
	if err, _ := callbackWork(ctx, 0); err != nil {
		t.Fatal(err)
	}
	walk := func() int {
		count := 0
		err := roots.Walk(ctx, RootsWalkRequest{
			OnMatch: func(callbackCtx context.Context, _ *SafeRootReference) (error, any) {
				if callbackCtx != ctx {
					t.Error("root callback received the detection cache context")
				}
				err, _ := callbackWork(callbackCtx, 0)
				count++
				return err, nil
			},
		})
		if err != nil {
			t.Fatal(err)
		}
		return count
	}
	if got := walk(); got != 0 {
		t.Fatalf("initial roots: %d", got)
	}
	writeFileForTest(t, filepath.Join(rootPath, "dyd/type"), "root")
	if got := walk(); got != 1 {
		t.Fatalf("roots after adding sentinel: %d", got)
	}
	if err := os.Remove(filepath.Join(rootPath, "dyd/type")); err != nil {
		t.Fatal(err)
	}
	if got := walk(); got != 0 {
		t.Fatalf("roots after removing sentinel: %d", got)
	}
	if calls != 1 {
		t.Fatalf("root callback lost parent cache reuse: %d executions", calls)
	}
}

func TestSharedIgnoreReaderInheritsCacheLayer(t *testing.T) {
	base := t.TempDir()
	ignorePath := filepath.Join(base, ".dyd-ignore")
	writeFileForTest(t, ignorePath, "ignored.txt\n")
	ctx := task.NewContext(context.Background(), 2)
	req := DydIgnoreRequest{BasePath: base, Path: base}
	err, first := readDydIgnore(ctx, req)
	if err != nil {
		t.Fatal(err)
	}
	writeFileForTest(t, ignorePath, "")
	for _, inherited := range []context.Context{ctx, task.WithConcurrency(ctx, 1)} {
		err, cached := readDydIgnore(inherited, req)
		if err != nil {
			t.Fatal(err)
		}
		if cached != first {
			t.Fatal("ignore reader did not reuse its cached matcher")
		}
	}
	err, fresh := readDydIgnore(task.WithCache(ctx), req)
	if err != nil {
		t.Fatal(err)
	}
	if fresh == first {
		t.Fatal("ignore reader reused the parent cache in a fresh layer")
	}
	path := dydfs.NewGlobPath(filepath.Join(base, "ignored.txt"), false)
	if err, ignored := first.Match(path); err != nil || !ignored {
		t.Fatalf("cached matcher: %v, %v", err, ignored)
	}
	if err, ignored := fresh.Match(path); err != nil || ignored {
		t.Fatalf("fresh matcher: %v, %v", err, ignored)
	}
}

func TestFileRequirementCopyInheritsCacheLayer(t *testing.T) {
	const hiddenFile = "dyd/assets/ignored/child.txt"
	base := t.TempDir()
	writeFileForTest(t, filepath.Join(base, hiddenFile), "content")
	writeFileForTest(t, filepath.Join(base, ".dyd-ignore"), "dyd/assets/ignored\n")
	ctx := task.NewContext(context.Background(), 3)
	copyIncludesFile := func(ctx context.Context) bool {
		dest := t.TempDir()
		if err := rootRequirementFileCopyDir(ctx, base, dest); err != nil {
			t.Fatal(err)
		}
		_, err := os.Stat(filepath.Join(dest, hiddenFile))
		if err != nil && !os.IsNotExist(err) {
			t.Fatal(err)
		}
		return err == nil
	}
	if copyIncludesFile(ctx) {
		t.Fatal("ignored file was copied")
	}
	writeFileForTest(t, filepath.Join(base, ".dyd-ignore"), "")
	if copyIncludesFile(ctx) {
		t.Fatal("copy replaced the inherited cache layer")
	}
	if !copyIncludesFile(task.WithCache(ctx)) {
		t.Fatal("copy did not use the fresh caller-provided cache layer")
	}
}
