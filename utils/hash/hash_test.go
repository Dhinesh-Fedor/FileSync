package hash

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHashFile(t *testing.T) {
	root := t.TempDir()

	path := filepath.Join(root, "test.txt")

	err := os.WriteFile(path, []byte("hello"), 0644)
	require.NoError(t, err)

	hash1, err := HashFile(path)
	require.NoError(t, err)

	hash2, err := HashFile(path)
	require.NoError(t, err)

	assert.Equal(t, hash1, hash2)
	assert.NotEmpty(t, hash1)
}