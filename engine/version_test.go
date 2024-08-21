package engine

import "testing"

func TestGetVersion(t *testing.T) {
	tineVersion = "v1.2.3"
	tineSha = "abcdef"
	timeVersion = "2020-01-01T00:00:00Z"
	goVersion = "1.14.2"

	v := GetVersion()
	if v.Major != 1 || v.Minor != 2 || v.Patch != 3 || v.GitSHA != "abcdef" {
		t.Errorf("unexpected version: %+v", v)
	}

	if got, want := DisplayVersion(), "1.2.3"; got != want {
		t.Errorf("unexpected version: %v", got)
	}

	if got, want := VersionString(), "1.2.3 abcdef 2020-01-01T00:00:00Z"; got != want {
		t.Errorf("unexpected version string: %v", got)
	}

	if got, want := BuildCompiler(), "1.14.2"; got != want {
		t.Errorf("unexpected compiler: %v", got)
	}

	if got, want := BuildTimestamp(), "2020-01-01T00:00:00Z"; got != want {
		t.Errorf("unexpected timestamp: %v", got)
	}
}
