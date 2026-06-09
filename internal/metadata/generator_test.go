package metadata

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"FileSync/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGeneratorBuild(t *testing.T) {
	root := t.TempDir()

	err := os.WriteFile(filepath.Join(root, "test.txt"), []byte("hello"), 0644)
	require.NoError(t, err)

	scan := models.ScannedFiles{
		Root:       root,
		Files:      []string{"test.txt"},
		FilesCount: 1,
		DirCount:   0,
	}

	generator := New(2)

	files, dirMeta, err := generator.Build(context.Background(), scan)

	require.NoError(t, err)
	require.Len(t, files, 1)

	file := files[0]

	assert.Equal(t, "test.txt", file.RelativePath)
	assert.NotEmpty(t, file.AbsolutePath)
	assert.Equal(t, int64(5), file.Size)
	assert.NotEmpty(t, file.SHA256)

	assert.Equal(t, 1, dirMeta.TotalFiles)
	assert.Equal(t, 0, dirMeta.TotalDirs)
	assert.Equal(t, int64(5), dirMeta.TotalSize)
	assert.False(t, dirMeta.LastScan.IsZero())
}

func TestGeneratorBuildCancelled(t *testing.T) {
	root := t.TempDir()

	err := os.WriteFile(filepath.Join(root, "test.txt"), []byte("hello"), 0644)
	require.NoError(t, err)

	scan := models.ScannedFiles{
		Root:       root,
		Files:      []string{"test.txt"},
		FilesCount: 1,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	generator := New(2)

	_, _, err = generator.Build(ctx, scan)

	require.Error(t, err)
}

func TestGeneratorBuildMultipleFiles(t *testing.T) {
	root := t.TempDir()

	err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("aaa"), 0644)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(root, "b.txt"), []byte("bbb"), 0644)
	require.NoError(t, err)

	scan := models.ScannedFiles{
		Root:       root,
		Files:      []string{"a.txt", "b.txt"},
		FilesCount: 2,
	}

	generator := New(4)

	files, _, err := generator.Build(context.Background(), scan)

	require.NoError(t, err)
	require.Len(t, files, 2)

	assert.Equal(t, "a.txt", files[0].RelativePath)
	assert.Equal(t, "b.txt", files[1].RelativePath)
}
