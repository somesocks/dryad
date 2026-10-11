package cli_test

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// Workers hold their execution slots until explicitly released. An atomic
// mkdir lock protects a shared active/peak counter, detecting oversubscription
// even if an extra worker starts and finishes between observations by the test.
const concurrencyRunScript = `g=$DYD_GARDEN
lock() {
    while ! mkdir "$g/counter-lock" 2>/dev/null; do sleep 0.01; done
}
lock
active=$(cat "$g/active")
active=$((active + 1))
printf '%s' "$active" > "$g/active"
peak=$(cat "$g/peak")
if [ "$active" -gt "$peak" ]; then printf '%s' "$active" > "$g/peak"; fi
rmdir "$g/counter-lock"
printf 'start:%s\n' "$id"
attempts=0
while [ ! -f "$g/release-$id" ]; do
    attempts=$((attempts + 1))
    if [ "$attempts" -gt 1000 ]; then exit 90; fi
    sleep 0.01
done
lock
active=$(cat "$g/active")
printf '%s' "$((active - 1))" > "$g/active"
rmdir "$g/counter-lock"
printf '%s' done > "$g/result-$id"
printf 'done:%s\n' "$id"
`

type runSession struct {
	events chan string
	done   chan struct{}
	err    error
	stderr bytes.Buffer
}

func startRunSession(t *testing.T, garden, input string, ids []string, args ...string) *runSession {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	args = append(args, "--scope=none", "--join-stdout=true")
	cmd := exec.CommandContext(ctx, runTestBinary, args...)
	cmd.Dir = garden
	cmd.Stdin = strings.NewReader(input)
	cmd.Env = append(os.Environ(), "GORACE=atexit_sleep_ms=0")
	session := &runSession{events: make(chan string, len(ids)*2+4), done: make(chan struct{})}
	// A real file avoids exec.Wait's output-copy goroutines inadvertently
	// waiting for abandoned children if the CLI itself returns too early.
	stderrPath := filepath.Join(garden, "run.stderr")
	stderrFile, err := os.Create(stderrPath)
	require.NoError(t, err)
	t.Cleanup(func() { stderrFile.Close() })
	cmd.Stderr = stderrFile
	stdout, err := cmd.StdoutPipe()
	require.NoError(t, err)
	require.NoError(t, cmd.Start())
	scanned := make(chan struct{})
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			session.events <- scanner.Text()
		}
		close(scanned)
	}()
	go func() {
		session.err = cmd.Wait()
		<-scanned
		content, _ := os.ReadFile(stderrPath)
		session.stderr.Write(content)
		close(session.done)
	}()
	t.Cleanup(func() {
		// Release every worker, including jobs not started yet, on assertion failure.
		for _, id := range ids {
			os.WriteFile(filepath.Join(garden, "release-"+id), nil, 0644)
		}
		select {
		case <-session.done:
		case <-time.After(15 * time.Second):
			cancel()
			<-session.done
		}
		cancel()
	})
	return session
}

func (session *runSession) nextEvent(t *testing.T) string {
	t.Helper()
	select {
	case event := <-session.events:
		return event
	case <-session.done:
		// Drain events already delivered before the process finished.
		select {
		case event := <-session.events:
			return event
		default:
			t.Fatalf("run ended before expected event: %v\n%s", session.err, session.stderr.String())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("workers did not start concurrently before release")
	}
	return ""
}

func TestRunConcurrency(t *testing.T) {
	for _, test := range []struct {
		name     string
		sprouts  int
		variants int
		parallel int
		args     []string
		stdin    bool
	}{
		{"default parallel", 1, 2, 0, []string{"sprout", "run", "dyd/sprouts/one"}, false},
		{"single sprout variants", 1, 2, 2, []string{"sprout", "run", "dyd/sprouts/one"}, false},
		{"single sprout serial", 1, 3, 1, []string{"sprout", "run", "dyd/sprouts/one"}, false},
		{"garden variants", 1, 4, 2, []string{"sprouts", "run"}, false},
		{"garden global limit", 2, 3, 3, []string{"sprouts", "run"}, false},
		{"garden serial", 2, 2, 1, []string{"sprouts", "run"}, false},
		{"stdin targets", 2, 1, 2, []string{"sprouts", "run", "--from-stdin"}, true},
		{"stdin variants", 1, 3, 2, []string{"sprouts", "run", "--from-stdin"}, true},
		{"stdin global limit", 2, 3, 3, []string{"sprouts", "run", "--from-stdin"}, true},
		{"stdin serial", 2, 2, 1, []string{"sprouts", "run", "--from-stdin"}, true},
	} {
		t.Run(test.name, func(t *testing.T) {
			garden := newRunGarden(t)
			writeRunFile(t, filepath.Join(garden, "active"), "0")
			writeRunFile(t, filepath.Join(garden, "peak"), "0")
			var ids []string
			var refs []string
			for i := 0; i < test.sprouts; i++ {
				name := []string{"one", "two"}[i]
				var variants []string
				for j := 0; j < test.variants; j++ {
					variant := string(rune('a' + j))
					variants = append(variants, variant)
					ids = append(ids, name+"-"+variant)
				}
				addRunSprout(t, garden, name, variants, concurrencyRunScript)
				refs = append(refs, "dyd/sprouts/"+name+"~flavor=any")
			}
			input := ""
			if test.stdin {
				input = strings.Join(refs, "\n") + "\n"
			}
			args := append([]string{}, test.args...)
			limit := test.parallel
			if limit > 0 {
				args = append(args, "--parallel="+strconv.Itoa(limit))
			} else {
				limit = 8
			}
			wantPeak := min(limit, len(ids))
			session := startRunSession(t, garden, input, ids, args...)
			started := map[string]bool{}
			// All available slots must be occupied while every worker is blocked.
			for i := 0; i < wantPeak; i++ {
				event := session.nextEvent(t)
				require.True(t, strings.HasPrefix(event, "start:"), event)
				require.False(t, started[event], "worker started twice")
				started[event] = true
			}
			for _, id := range ids {
				writeRunFile(t, filepath.Join(garden, "release-"+id), "")
			}
			select {
			case <-session.done:
			case <-time.After(10 * time.Second):
				t.Fatal("run did not wait for and finish released workers")
			}
			require.NoError(t, session.err, session.stderr.String())
			for _, id := range ids {
				assertRunResult(t, garden, id, "done")
			}
			peak, err := os.ReadFile(filepath.Join(garden, "peak"))
			require.NoError(t, err)
			require.Equal(t, strconv.Itoa(wantPeak), string(peak), "invocation-wide concurrency limit")
			active, err := os.ReadFile(filepath.Join(garden, "active"))
			require.NoError(t, err)
			require.Equal(t, "0", string(active))
		})
	}
}

func TestRunFailureWaitsForStartedVariants(t *testing.T) {
	garden := newRunGarden(t)
	addRunSprout(t, garden, "one", []string{"a", "b"}, `printf 'start:%s\n' "$id"
attempts=0
while [ ! -f "$DYD_GARDEN/release-$id" ]; do
    attempts=$((attempts + 1))
    if [ "$attempts" -gt 1000 ]; then exit 90; fi
    sleep 0.01
done
if [ "$id" = one-a ]; then exit 7; fi
printf '%s' done > "$DYD_GARDEN/result-$id"
`)
	session := startRunSession(t, garden, "", []string{"one-a", "one-b"}, "sprout", "run", "dyd/sprouts/one", "--parallel=2")
	starts := []string{session.nextEvent(t), session.nextEvent(t)}
	require.ElementsMatch(t, []string{"start:one-a", "start:one-b"}, starts)
	// The sibling remains blocked when the failing variant exits.
	writeRunFile(t, filepath.Join(garden, "release-one-a"), "")
	deadline := time.After(5 * time.Second)
	for {
		log, err := os.ReadFile(filepath.Join(garden, "run.stderr"))
		require.NoError(t, err)
		if strings.Contains(string(log), "sprout threw error during execution") {
			break
		}
		select {
		case <-session.done:
			t.Fatal("run returned while a started sibling was still blocked")
		case <-deadline:
			t.Fatal("variant failure was not reported")
		case <-time.After(10 * time.Millisecond):
		}
	}
	select {
	case <-session.done:
		t.Fatal("run returned while a started sibling was still blocked")
	default:
	}
	writeRunFile(t, filepath.Join(garden, "release-one-b"), "")
	select {
	case <-session.done:
	case <-time.After(10 * time.Second):
		t.Fatal("run did not complete after variant failure")
	}
	require.Error(t, session.err)
	require.Contains(t, session.stderr.String(), "error running sprout variant flavor=a")
	assertRunResult(t, garden, "one-b", "done")
}

func TestRunParallelIgnoreErrorsContinuesOtherSprouts(t *testing.T) {
	garden := newRunGarden(t)
	ids := []string{"one-a", "one-b", "two-a", "two-b"}
	for _, name := range []string{"one", "two"} {
		addRunSprout(t, garden, name, []string{"a", "b"}, `printf 'start:%s\n' "$id"
attempts=0
while [ ! -f "$DYD_GARDEN/release-$id" ]; do
    attempts=$((attempts + 1))
    if [ "$attempts" -gt 1000 ]; then exit 90; fi
    sleep 0.01
done
if [ "$id" = one-a ]; then exit 7; fi
printf '%s' done > "$DYD_GARDEN/result-$id"
`)
	}
	session := startRunSession(t, garden, "dyd/sprouts/one\ndyd/sprouts/two\n", ids,
		"sprouts", "run", "--from-stdin", "--parallel=2", "--ignore-errors")
	starts := []string{session.nextEvent(t), session.nextEvent(t)}
	require.ElementsMatch(t, []string{"start:one-a", "start:one-b"}, starts)
	for _, id := range ids {
		writeRunFile(t, filepath.Join(garden, "release-"+id), "")
	}
	select {
	case <-session.done:
	case <-time.After(10 * time.Second):
		t.Fatal("ignored variant error prevented remaining sprouts from finishing")
	}
	require.NoError(t, session.err, session.stderr.String())
	require.Contains(t, session.stderr.String(), "error running sprout variant flavor=a")
	for _, id := range []string{"one-b", "two-a", "two-b"} {
		assertRunResult(t, garden, id, "done")
	}
	for _, line := range strings.Split(session.stderr.String(), "\n") {
		if strings.Contains(line, "sprout run completed") {
			require.NotContains(t, line, "dyd/sprouts/one", "failed sprouts must not be logged as completed")
		}
	}
}

func TestRunSerialFailureAndIgnoreErrors(t *testing.T) {
	for _, ignore := range []bool{false, true} {
		t.Run(fmt.Sprintf("ignore=%v", ignore), func(t *testing.T) {
			garden := newRunGarden(t)
			for _, name := range []string{"one", "two"} {
				addRunSprout(t, garden, name, []string{"a", "b"}, `if [ "$id" = one-a ]; then exit 7; fi
printf '%s' done > "$DYD_GARDEN/result-$id"
`)
			}
			args := []string{"sprouts", "run", "--from-stdin", "--parallel=1"}
			if ignore {
				args = append(args, "--ignore-errors")
			}
			_, stderr, err := runCommand(t, garden, "dyd/sprouts/one\ndyd/sprouts/two\n", args...)
			require.Contains(t, stderr, "error running sprout variant flavor=a")
			require.NoFileExists(t, filepath.Join(garden, "result-one-b"))
			if ignore {
				require.NoError(t, err, stderr)
				assertRunResult(t, garden, "two-a", "done")
				assertRunResult(t, garden, "two-b", "done")
			} else {
				require.Error(t, err)
				require.NoFileExists(t, filepath.Join(garden, "result-two-a"))
			}
		})
	}
}
