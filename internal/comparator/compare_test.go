package comparator

import (
	"testing"

	"FileSync/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestCompare(t *testing.T) {
	source := []models.FileInfo{
		{
			RelativePath: "a.txt",
			Size:         10,
			SHA256:       "hash1",
		},
		{
			RelativePath: "b.txt",
			Size:         20,
			SHA256:       "hash2",
		},
	}

	destination := []models.FileInfo{
		{
			RelativePath: "b.txt",
			Size:         999,
			SHA256:       "different",
		},
		{
			RelativePath: "c.txt",
			Size:         30,
			SHA256:       "hash3",
		},
	}

	result := Compare(source, destination)

	assert.Equal(t, 1, result.Added)
	assert.Equal(t, 1, result.Modified)
	assert.Equal(t, 1, result.Deleted)

	assert.Len(t, result.Changes, 3)
}