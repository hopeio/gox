package fs

import (
	"os"
	"path/filepath"
	"testing"
)

// TestDedupSkipsEmptyFiles 空文件 MD5 恒等，不应互相判定为重复而被删除。
func TestDedupSkipsEmptyFiles(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), nil, 0644); err != nil {
			t.Fatal(err)
		}
	}
	var removed []string
	err := DirsDuplicateHandle(func(path1, path2 string) error {
		removed = append(removed, path2)
		return nil
	}, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 0 {
		t.Fatalf("empty files treated as duplicates: %v", removed)
	}
}

// TestDedupStillFindsRealDuplicates 非空同内容文件仍要被判为重复。
func TestDedupStillFindsRealDuplicates(t *testing.T) {
	dir := t.TempDir()
	for _, name := range []string{"a.txt", "b.txt"} {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("same content"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	var removed []string
	err := DirsDuplicateHandle(func(path1, path2 string) error {
		removed = append(removed, path2)
		return nil
	}, dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(removed) != 1 {
		t.Fatalf("expected exactly 1 duplicate, got %v", removed)
	}
}
