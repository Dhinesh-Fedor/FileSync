package synchronizer

import (
	"os"
	"path/filepath"
	"testing"

	"FileSync/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestAddedFile(t *testing.T) {

	source := t.TempDir()
	destination := t.TempDir()

	err := os.WriteFile(
		filepath.Join(source, "test.txt"),
		[]byte("hello"),
		0644,
	)

	assert.NoError(t, err)

	result := models.CompareResult{
		Changes: []models.FileChange{
			{
				Type: models.Added,
				RelativePath: "test.txt",
			},
		},
	}

	report, err := Sync(
		source,
		destination,
		result,
	)

	assert.NoError(t, err)

	assert.Equal(t, 1, report.Added)

	_, err = os.Stat(
		filepath.Join(destination, "test.txt"),
	)

	assert.NoError(t, err)
}


func TestModifiedFile(t *testing.T) {

	source := t.TempDir()
	destination := t.TempDir()

	err := os.WriteFile(
		filepath.Join(source, "test.txt"),
		[]byte("new"),
		0644,
	)

	assert.NoError(t, err)

	err = os.WriteFile(
		filepath.Join(destination, "test.txt"),
		[]byte("old"),
		0644,
	)

	assert.NoError(t, err)

	result := models.CompareResult{
		Changes: []models.FileChange{
			{
				Type: models.Modified,
				RelativePath: "test.txt",
			},
		},
	}

	report, err := Sync(
		source,
		destination,
		result,
	)

	assert.NoError(t, err)

	assert.Equal(t, 1, report.Modified)
}


func TestDeletedFile(t *testing.T) {

	source := t.TempDir()
	destination := t.TempDir()

	file := filepath.Join(
		destination,
		"delete.txt",
	)

	err := os.WriteFile(
		file,
		[]byte("data"),
		0644,
	)

	assert.NoError(t, err)

	result := models.CompareResult{
		Changes: []models.FileChange{
			{
				Type: models.Deleted,
				RelativePath: "delete.txt",
			},
		},
	}

	report, err := Sync(
		source,
		destination,
		result,
	)

	assert.NoError(t, err)

	assert.Equal(t, 1, report.Deleted)

	_, err = os.Stat(file)

	assert.True(t, os.IsNotExist(err))
}
