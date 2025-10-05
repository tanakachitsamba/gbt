package main

import (
	"os/exec"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestEncode(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	var capturedArgs []string
	execCommand = func(name string, arg ...string) *exec.Cmd {
		capturedArgs = append([]string{name}, arg...)
		return exec.Command("bash", "-c", `printf '{"tokens":[1,2],"count":2}'`)
	}

	output, err := encode("hello world")
	require.NoError(t, err)
	require.Equal(t, []int{1, 2}, output.Tokens)
	require.Equal(t, 2, output.Count)
	require.Equal(t, []string{"node", "encoder.js", "hello world"}, capturedArgs)
}

func TestEncode_CommandError(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("bash", "-c", `echo error 1>&2; exit 1`)
	}

	_, err := encode("fail")
	require.Error(t, err)
}

func TestEncode_InvalidJSON(t *testing.T) {
	original := execCommand
	defer func() { execCommand = original }()

	execCommand = func(name string, arg ...string) *exec.Cmd {
		return exec.Command("bash", "-c", `printf 'not json'`)
	}

	_, err := encode("fail")
	require.Error(t, err)
}
