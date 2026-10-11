package cli_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

var runTestBinary string
var runFixtureBinary string

// Exercise the real CLI and subprocess streams, including race detection in the
// scheduler. Build once so each case can use an isolated miniature garden.
func TestMain(m *testing.M) {
	dir, err := os.MkdirTemp("", "dryad-run-tests-")
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	runTestBinary = filepath.Join(dir, "dryad")
	buildArgs := []string{"build", "-o", runTestBinary}
	if runTestRace {
		buildArgs = append(buildArgs, "-race")
	}
	buildArgs = append(buildArgs, "..")
	build := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), buildArgs...)
	build.Stdout, build.Stderr = os.Stdout, os.Stderr
	if err := build.Run(); err != nil {
		os.RemoveAll(dir)
		os.Exit(1)
	}
	runFixtureBinary = runTestBinary
	if runTestRace {
		// Legacy fixtures build roots as well as running sprouts. Root-build
		// cleanup has unrelated existing races in filesystem.RemoveAll; keep
		// those fixtures on a regular binary and instrument the run-only tests.
		plainDir := filepath.Join(dir, "plain")
		if err := os.Mkdir(plainDir, 0755); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.RemoveAll(dir)
			os.Exit(1)
		}
		runFixtureBinary = filepath.Join(plainDir, "dryad")
		build := exec.Command(filepath.Join(runtime.GOROOT(), "bin", "go"), "build", "-o", runFixtureBinary, "..")
		build.Stdout, build.Stderr = os.Stdout, os.Stderr
		if err := build.Run(); err != nil {
			os.RemoveAll(dir)
			os.Exit(1)
		}
	}
	code := m.Run()
	os.RemoveAll(dir)
	os.Exit(code)
}

func writeRunFile(t *testing.T, path, content string) {
	t.Helper()
	require.NoError(t, os.MkdirAll(filepath.Dir(path), 0755))
	require.NoError(t, os.WriteFile(path, []byte(content), 0755))
}

func newRunGarden(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	writeRunFile(t, filepath.Join(dir, "dyd", "type"), "garden")
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "dyd", "sprouts"), 0755))
	return dir
}

func addRunSprout(t *testing.T, garden, name string, variants []string, script string) {
	t.Helper()
	sprout := filepath.Join(garden, "dyd", "heap", "sprouts", name)
	writeRunFile(t, filepath.Join(sprout, "dyd", "type"), "sprout")
	require.NoError(t, os.MkdirAll(filepath.Join(sprout, "dyd", "dependencies"), 0755))
	for _, variant := range variants {
		id := name + "-" + variant
		stem := filepath.Join(garden, "dyd", "heap", "stems", id)
		writeRunFile(t, filepath.Join(stem, "dyd", "type"), "stem")
		writeRunFile(t, filepath.Join(stem, "dyd", "commands", "dyd-stem-run"), "#!/bin/sh\nset -eu\nid='"+id+"'\n"+script)
		require.NoError(t, os.MkdirAll(filepath.Join(stem, "dyd", "dependencies"), 0755))
		dependency := "stem"
		if variant != "" {
			dependency += "~flavor=" + variant
			writeRunFile(t, filepath.Join(stem, "dyd", "traits", "flavor"), variant)
		}
		require.NoError(t, os.Symlink(stem, filepath.Join(sprout, "dyd", "dependencies", dependency)))
	}
	require.NoError(t, os.Symlink(sprout, filepath.Join(garden, "dyd", "sprouts", name)))
}

func runCommand(t *testing.T, garden, input string, args ...string) (string, string, error) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if args[0] != "stem" {
		args = append(args, "--scope=none")
	}
	cmd := exec.CommandContext(ctx, runTestBinary, args...)
	cmd.Dir = garden
	cmd.Stdin = strings.NewReader(input)
	cmd.Env = append(os.Environ(), "GORACE=atexit_sleep_ms=0")
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	err := cmd.Run()
	require.NoError(t, ctx.Err(), "command timed out: %s", stderr.String())
	return stdout.String(), stderr.String(), err
}

const stdinRunScript = `line=EOF
IFS= read -r line || line=EOF
printf '%s' "$line" > "$DYD_GARDEN/result-$id"
printf 'stdout-%s\n' "$id"
printf 'stderr-%s\n' "$id" >&2
`

func assertRunResult(t *testing.T, garden, id, expected string) {
	t.Helper()
	content, err := os.ReadFile(filepath.Join(garden, "result-"+id))
	require.NoError(t, err)
	require.Equal(t, expected, string(content))
}

func TestSproutRunStdinModes(t *testing.T) {
	for _, test := range []struct {
		name     string
		variants []string
		options  []string
		want     []string
	}{
		{"single auto", []string{"a"}, nil, []string{"first"}},
		{"single inherit", []string{"a"}, []string{"--join-stdin=true"}, []string{"first"}},
		{"single none", []string{"a"}, []string{"--join-stdin=false"}, []string{"EOF"}},
		{"multi auto", []string{"a", "b"}, []string{"--parallel=4"}, []string{"EOF", "EOF"}},
		{"serial auto", []string{"a", "b"}, []string{"--parallel=1"}, []string{"first", "second"}},
		{"serial inherit", []string{"a", "b"}, []string{"--parallel=1", "--join-stdin=true"}, []string{"first", "second"}},
		{"serial none", []string{"a", "b"}, []string{"--parallel=1", "--join-stdin=false"}, []string{"EOF", "EOF"}},
	} {
		t.Run(test.name, func(t *testing.T) {
			garden := newRunGarden(t)
			addRunSprout(t, garden, "one", test.variants, stdinRunScript)
			args := append([]string{"sprout", "run", "dyd/sprouts/one"}, test.options...)
			_, stderr, err := runCommand(t, garden, "first\nsecond\n", args...)
			require.NoError(t, err, stderr)
			for i, variant := range test.variants {
				assertRunResult(t, garden, "one-"+variant, test.want[i])
			}
		})
	}
}

func TestSproutRunStdinUsesSelectedVariants(t *testing.T) {
	garden := newRunGarden(t)
	addRunSprout(t, garden, "one", []string{"a", "b"}, stdinRunScript)
	_, stderr, err := runCommand(t, garden, "selected\n", "sprout", "run", "dyd/sprouts/one~flavor=a", "--join-stdin=true", "--parallel=4")
	require.NoError(t, err, stderr)
	assertRunResult(t, garden, "one-a", "selected")
	require.NoFileExists(t, filepath.Join(garden, "result-one-b"))
}

func TestSproutsRunStdinIsInvocationWide(t *testing.T) {
	for _, options := range [][]string{{"--parallel=4"}, {"--parallel=4", "--join-stdin=false"}} {
		garden := newRunGarden(t)
		addRunSprout(t, garden, "one", []string{"a"}, stdinRunScript)
		addRunSprout(t, garden, "two", []string{"a"}, stdinRunScript)
		stdout, stderr, err := runCommand(t, garden, "not-for-children\n", append([]string{"sprouts", "run"}, options...)...)
		require.NoError(t, err, stderr)
		assertRunResult(t, garden, "one-a", "EOF")
		assertRunResult(t, garden, "two-a", "EOF")
		require.Empty(t, stdout)
		require.NotContains(t, stderr, "stderr-one-a")
	}
	garden := newRunGarden(t)
	addRunSprout(t, garden, "one", []string{"a"}, stdinRunScript)
	addRunSprout(t, garden, "two", []string{"a"}, stdinRunScript)
	_, stderr, err := runCommand(t, garden, "filtered\n", "sprouts", "run", "--include=dyd/sprouts/one", "--join-stdin=true", "--parallel=4")
	require.NoError(t, err, stderr)
	assertRunResult(t, garden, "one-a", "filtered")
	require.NoFileExists(t, filepath.Join(garden, "result-two-a"))
}

func TestSproutsRunSerialStdin(t *testing.T) {
	garden := newRunGarden(t)
	addRunSprout(t, garden, "one", []string{"a"}, stdinRunScript)
	addRunSprout(t, garden, "two", []string{"a"}, stdinRunScript)
	_, stderr, err := runCommand(t, garden, "first\nsecond\n", "sprouts", "run", "--parallel=1", "--join-stdin=true")
	require.NoError(t, err, stderr)
	var results []string
	for _, id := range []string{"one-a", "two-a"} {
		content, err := os.ReadFile(filepath.Join(garden, "result-"+id))
		require.NoError(t, err)
		results = append(results, string(content))
	}
	require.ElementsMatch(t, []string{"first", "second"}, results)
}

func TestStemRunStdinModes(t *testing.T) {
	for _, options := range [][]string{nil, {"--join-stdin=true"}, {"--join-stdin=false"}} {
		garden := newRunGarden(t)
		addRunSprout(t, garden, "one", []string{"a"}, stdinRunScript)
		args := append([]string{"stem", "run", "dyd/heap/stems/one-a"}, options...)
		_, stderr, err := runCommand(t, garden, "stem-input\n", args...)
		require.NoError(t, err, stderr)
		want := "stem-input"
		if len(options) > 0 && options[0] == "--join-stdin=false" {
			want = "EOF"
		}
		assertRunResult(t, garden, "one-a", want)
	}
}

func TestSproutsRunFromStdinDoesNotJoinInput(t *testing.T) {
	for _, options := range [][]string{nil, {"--join-stdin=false"}, {"--parallel=1"}} {
		garden := newRunGarden(t)
		addRunSprout(t, garden, "one", []string{"a", "b"}, stdinRunScript)
		_, stderr, err := runCommand(t, garden, "dyd/sprouts/one~flavor=any\n", append([]string{"sprouts", "run", "--from-stdin"}, options...)...)
		require.NoError(t, err, stderr)
		assertRunResult(t, garden, "one-a", "EOF")
		assertRunResult(t, garden, "one-b", "EOF")
	}
}

func TestRunRejectsStdinConflictsBeforeStartingChildren(t *testing.T) {
	for _, args := range [][]string{
		{"sprout", "run", "dyd/sprouts/one", "--parallel=4", "--join-stdin=true"},
		{"sprouts", "run", "--parallel=4", "--join-stdin=true"},
		{"sprouts", "run", "--from-stdin", "--parallel=1", "--join-stdin=true"},
	} {
		garden := newRunGarden(t)
		addRunSprout(t, garden, "one", []string{"a", "b"}, stdinRunScript)
		_, stderr, err := runCommand(t, garden, "dyd/sprouts/one\n", args...)
		require.Error(t, err)
		require.Contains(t, stderr, "--join-stdin=true")
		require.NoFileExists(t, filepath.Join(garden, "result-one-a"))
		require.NoFileExists(t, filepath.Join(garden, "result-one-b"))
	}
	// Separate single-variant sprouts must not each decide they own stdin.
	garden := newRunGarden(t)
	addRunSprout(t, garden, "one", []string{"a"}, stdinRunScript)
	addRunSprout(t, garden, "two", []string{"a"}, stdinRunScript)
	_, stderr, err := runCommand(t, garden, "input\n", "sprouts", "run", "--join-stdin=true", "--parallel=2")
	require.Error(t, err)
	require.Contains(t, stderr, "requires --parallel=1")
	require.NoFileExists(t, filepath.Join(garden, "result-one-a"))
	require.NoFileExists(t, filepath.Join(garden, "result-two-a"))
}

func TestRunOutputJoiningAndVariantLogFiles(t *testing.T) {
	garden := newRunGarden(t)
	addRunSprout(t, garden, "one", []string{"a", "b"}, stdinRunScript)
	stdout, stderr, err := runCommand(t, garden, "", "sprout", "run", "dyd/sprouts/one", "--parallel=4")
	require.NoError(t, err, stderr)
	require.Contains(t, stdout, "stdout-one-a")
	require.Contains(t, stderr, "stderr-one-b")
	stdout, stderr, err = runCommand(t, garden, "", "sprout", "run", "dyd/sprouts/one", "--join-stdout=false", "--join-stderr=false")
	require.NoError(t, err, stderr)
	require.Empty(t, stdout)
	require.NotContains(t, stderr, "stderr-one-a")
	stdout, stderr, err = runCommand(t, garden, "", "sprouts", "run", "--join-stdout=true", "--join-stderr=true")
	require.NoError(t, err, stderr)
	require.Contains(t, stdout, "stdout-one-a")
	require.Contains(t, stderr, "stderr-one-b")
	logs := filepath.Join(garden, "logs")
	require.NoError(t, os.Mkdir(logs, 0755))
	stdout, stderr, err = runCommand(t, garden, "", "sprout", "run", "dyd/sprouts/one", "--join-stdout=true", "--join-stderr=true", "--log-stdout="+logs, "--log-stderr="+logs)
	require.NoError(t, err, stderr)
	require.Empty(t, stdout)
	require.NotContains(t, stderr, "stderr-one-a")
	for _, variant := range []string{"a", "b"} {
		base := "dyd-sprout-run--dyd-sprouts-one--flavor=" + variant
		out, err := os.ReadFile(filepath.Join(logs, base+".out"))
		require.NoError(t, err)
		require.Equal(t, "stdout-one-"+variant+"\n", string(out))
		errOut, err := os.ReadFile(filepath.Join(logs, base+".err"))
		require.NoError(t, err)
		require.Equal(t, "stderr-one-"+variant+"\n", string(errOut))
	}
}

// Reuse garden fixtures from the Dryad integration suite against this freshly
// built CLI, without depending on potentially stale built test sprouts.
func TestExistingRunIntegrationFixtures(t *testing.T) {
	for _, name := range []string{
		"sprout-run-02--supports-executables-and-log",
		"sprout-run-03--join-and-log-files",
		"sprout-run-04--fails-when-the-sprout",
		"sprout-run-08--variant-any-runs-set",
		"sprout-run-09--qualified-variant-ref",
		"sprout-run-10--env-requirements-runtime",
		"sprouts-run-02--variant-any-runs-set",
		"sprouts-run-03--qualified-stdin-refs",
		"stem-run-02--executes-a-stem-main",
		"root-build-02--basic-root",
		"root-develop-16--root-develop-shell-command",
	} {
		t.Run(name, func(t *testing.T) {
			root := filepath.Join("..", "..", "..", "..", "tests", name)
			stem := t.TempDir()
			require.NoError(t, os.MkdirAll(filepath.Join(stem, "dyd"), 0755))
			assets := filepath.Join(root, "dyd", "assets", "assets")
			if _, err := os.Stat(assets); err == nil {
				require.NoError(t, os.CopyFS(filepath.Join(stem, "dyd", "assets"), os.DirFS(assets)))
			} else {
				require.True(t, os.IsNotExist(err), "unexpected assets error: %v", err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, "/bin/bash", filepath.Join(root, "dyd", "assets", "commands", "dyd-stem-run"))
			cmd.Env = append(os.Environ(), "DYD_STEM="+stem, "DYD_LOG_LEVEL=info", "PATH="+filepath.Dir(runFixtureBinary)+string(os.PathListSeparator)+os.Getenv("PATH"))
			output, err := cmd.CombinedOutput()
			require.NoError(t, ctx.Err(), "fixture timed out: %s", output)
			require.NoError(t, err, string(output))
		})
	}
}
