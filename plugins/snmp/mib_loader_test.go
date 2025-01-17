package snmp

import (
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
)

type testMibLoader struct {
	folders []string
	files   []string
}

func (t *testMibLoader) appendPath(path string) {
	t.folders = append(t.folders, path)
}

func (t *testMibLoader) loadModule(path string) error {
	t.files = append(t.files, path)
	return nil
}

func TestFolderLookup(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("skipping test on windows")
	}

	tests := []struct {
		name    string
		mibPath [][]string
		paths   [][]string
		files   []string
	}{
		{
			name:    "loading folders",
			mibPath: [][]string{{"testdata", "loadMibsFromPath", "root"}},
			paths: [][]string{
				{"testdata", "loadMibsFromPath", "root"},
				{"testdata", "loadMibsFromPath", "root", "dirOne"},
				{"testdata", "loadMibsFromPath", "root", "dirOne", "dirTwo"},
				{"testdata", "loadMibsFromPath", "linkTarget"},
			},
			files: []string{"empty", "emptyFile"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			loader := &testMibLoader{}

			var givenPath []string
			for _, paths := range tt.mibPath {
				rootPath := filepath.Join(paths...)
				givenPath = append(givenPath, rootPath)
			}

			err := LoadMibsFromPath(givenPath, loader)
			require.NoError(t, err)

			var folders []string
			for _, pathSlice := range tt.paths {
				path := filepath.Join(pathSlice...)
				folders = append(folders, path)
			}
			require.Equal(t, folders, loader.folders)
			require.Equal(t, tt.files, loader.files)
		})
	}
}

func TestMissingMibPath(t *testing.T) {
	path := []string{"non-existing-directory"}
	require.NoError(t, LoadMibsFromPath(path, &GosmiMibLoader{}))
}

func BenchmarkMibLoading(b *testing.B) {
	path := []string{"testdata/gosmi"}
	for i := 0; i < b.N; i++ {
		require.NoError(b, LoadMibsFromPath(path, &GosmiMibLoader{}))
	}
}
