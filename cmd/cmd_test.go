package cmd

import (
	"bytes"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestExecuteCommand(t *testing.T) {
	cmd := NewCmd()

	b := bytes.NewBufferString("")
	cmd.SetOut(b)
	cmd.SetArgs([]string{"--help"})
	cmd.Execute()

	out, err := io.ReadAll(b)
	if err != nil {
		t.Fatal(err)
	}
	bin, _ := os.ReadFile("./testdata/help.txt")
	bin = bytes.ReplaceAll(bin, []byte{'\r'}, nil)
	result := strings.ReplaceAll(string(out), "\r", "")
	require.Equal(t, string(bin), result)
}
