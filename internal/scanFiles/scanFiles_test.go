package scanFiles

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScanFiles(t *testing.T) {
	root := t.TempDir()

	err := os.WriteFile(filepath.Join(root, "a.txt"), []byte("hello"), 0644)
	require.NoError(t, err)

	err = os.Mkdir(filepath.Join(root, "docs"), 0755)
	require.NoError(t, err)

	err = os.WriteFile(filepath.Join(root, "docs", "b.txt"), []byte("world"), 0644)
	require.NoError(t, err)

	result, err := Scanner(context.Background(), root)

	require.NoError(t, err)

	assert.Equal(t, 2, result.FilesCount)
	assert.Equal(t, 1, result.DirCount)
	assert.Len(t, result.Files, 2)

	assert.Contains(t, result.Files, "a.txt")
	assert.Contains(t, result.Files, "docs/b.txt")
}