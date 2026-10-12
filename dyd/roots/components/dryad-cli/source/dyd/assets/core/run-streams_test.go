package core

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestInputConfigResolve(t *testing.T) {
	for _, mode := range []InputMode{InputAuto, InputNone, InputInherit} {
		for _, count := range []int{0, 1, 2} {
			for _, parallel := range []int{1, 4} {
				for _, fromStdin := range []bool{false, true} {
					resolved, err := (InputConfig{Mode: mode}).Resolve(count, parallel, fromStdin)
					conflict := fromStdin || (count > 1 && parallel > 1)
					if mode == InputInherit && conflict {
						require.Error(t, err, "mode=%v count=%v parallel=%v fromStdin=%v", mode, count, parallel, fromStdin)
						continue
					}
					want := mode
					if mode == InputAuto {
						want = InputInherit
						if conflict {
							want = InputNone
						}
					}
					require.NoError(t, err)
					require.Equal(t, want, resolved.Mode, "mode=%v count=%v parallel=%v fromStdin=%v", mode, count, parallel, fromStdin)
				}
			}
		}
	}
	_, err := (InputConfig{Mode: InputMode(99)}).Resolve(1, 1, false)
	require.Error(t, err)
	_, err = (InputConfig{}).Resolve(-1, 1, false)
	require.Error(t, err)
	_, err = (InputConfig{}).Resolve(1, 0, false)
	require.Error(t, err)
}

func TestStemRunCommandRejectsInvalidStreamConfigBeforePreparation(t *testing.T) {
	for _, test := range []struct {
		name    string
		request StemRunRequest
		message string
	}{
		{"input mode", StemRunRequest{Stdin: InputConfig{Mode: InputMode(99)}}, "invalid input mode"},
		{"output mode", StemRunRequest{Stdout: OutputConfig{Mode: OutputMode(99)}}, "stdout: invalid output mode"},
		{"missing log path", StemRunRequest{Stderr: OutputConfig{Mode: OutputLog}}, "stderr: output log mode requires a directory"},
		{"unused log path", StemRunRequest{Stdout: OutputConfig{Mode: OutputNone, Path: "logs"}}, "stdout: output path and name require log mode"},
	} {
		t.Run(test.name, func(t *testing.T) {
			instance, err := StemRunCommand(test.request)
			require.Nil(t, instance)
			require.Error(t, err)
			require.Contains(t, err.Error(), test.message)
		})
	}
}
