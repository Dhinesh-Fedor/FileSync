package metadata

import (
	"os"
	"path/filepath"
	"testing"

	"FileSync/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuild(t *testing.T) {

	root := t.TempDir()

	err := os.WriteFile(filepath.Join(root, "test.txt"), []byte("hello"), 0644)
	require.NoError(t, err)

	scan := models.ScannedFiles{
		Root:       root,
		Files:      []string{"test.txt"},
		FilesCount: 1,
		DirCount:   0,
	}

	files, dirMeta, err := Build(scan)

	require.NoError(t, err)

	require.Len(t, files, 1)

	file := files[0]

	assert.Equal(t,	"test.txt", file.RelativePath)
	assert.NotEmpty(t, file.AbsolutePath)
	assert.Equal(t, int64(5), file.Size)
	assert.NotEmpty(t, file.SHA256)
	assert.Equal(t,	1, dirMeta.TotalFiles)
	assert.Equal(t,	0, dirMeta.TotalDirs)
	assert.Equal(t,	int64(5), dirMeta.TotalSize)
	assert.False(t, dirMeta.LastScan.IsZero())
}