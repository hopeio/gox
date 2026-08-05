package zip

import (
	"archive/zip"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCompressDir(t *testing.T) {
	src := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(src, "a.txt"), []byte("hello"), 0o644))
	require.NoError(t, os.Mkdir(filepath.Join(src, "sub"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(src, "sub", "b.txt"), []byte("world"), 0o644))

	t.Run("without root dir", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "out.zip")
		require.NoError(t, CompressDir(src, out, false))

		names := zipNames(t, out)
		require.Contains(t, names, "a.txt")
		require.Contains(t, names, "sub/")
		require.Contains(t, names, "sub/b.txt")
		require.NotContains(t, names, filepath.Base(src)+"/")
		require.Equal(t, []byte("hello"), zipFileContent(t, out, "a.txt"))
		require.Equal(t, []byte("world"), zipFileContent(t, out, "sub/b.txt"))
	})

	t.Run("with root dir", func(t *testing.T) {
		out := filepath.Join(t.TempDir(), "out.zip")
		require.NoError(t, CompressDir(src, out, true))

		root := filepath.Base(src)
		names := zipNames(t, out)
		require.Contains(t, names, root+"/")
		require.Contains(t, names, root+"/a.txt")
		require.Contains(t, names, root+"/sub/")
		require.Contains(t, names, root+"/sub/b.txt")
		require.Equal(t, []byte("hello"), zipFileContent(t, out, root+"/a.txt"))
	})
}

func TestCompressDir_EmptyDir(t *testing.T) {
	src := t.TempDir()
	out := filepath.Join(t.TempDir(), "empty.zip")
	require.NoError(t, CompressDir(src, out, false))

	names := zipNames(t, out)
	require.Empty(t, names)

	out2 := filepath.Join(t.TempDir(), "empty_root.zip")
	require.NoError(t, CompressDir(src, out2, true))
	names2 := zipNames(t, out2)
	require.Equal(t, []string{filepath.Base(src) + "/"}, names2)
}

func TestCompressDir_SourceNotExist(t *testing.T) {
	out := filepath.Join(t.TempDir(), "x.zip")
	err := CompressDir(filepath.Join(t.TempDir(), "missing"), out, false)
	require.Error(t, err)
}

func TestCompressDir_TargetInvalid(t *testing.T) {
	src := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(src, "a.txt"), []byte("x"), 0o644))
	err := CompressDir(src, filepath.Join(t.TempDir(), "nope", "out.zip"), false)
	require.Error(t, err)
}

func zipNames(t *testing.T, zipPath string) []string {
	t.Helper()
	r, err := zip.OpenReader(zipPath)
	require.NoError(t, err)
	defer r.Close()
	names := make([]string, 0, len(r.File))
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	return names
}

func zipFileContent(t *testing.T, zipPath, name string) []byte {
	t.Helper()
	r, err := zip.OpenReader(zipPath)
	require.NoError(t, err)
	defer r.Close()
	for _, f := range r.File {
		if f.Name != name {
			continue
		}
		rc, err := f.Open()
		require.NoError(t, err)
		defer rc.Close()
		data, err := io.ReadAll(rc)
		require.NoError(t, err)
		return data
	}
	t.Fatalf("file %q not found in %s", name, zipPath)
	return nil
}
