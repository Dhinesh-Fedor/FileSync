package pairmanager

import (
	"os"
	"testing"

	"FileSync/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestAddPair(t *testing.T) {
	root := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	_ = os.Remove(ConfigFile)

	pair, err := Add(
		"/tmp/source",
		"/tmp/destination",
	)

	assert.NoError(t, err)

	assert.Equal(t, 1, pair.ID)
	assert.Equal(t, "/tmp/source", pair.Source)
	assert.Equal(t, "/tmp/destination", pair.Destination)
}

func TestGetPair(t *testing.T) {
	root := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	_ = os.Remove(ConfigFile)

	pair, _ := Add(
		"/tmp/source",
		"/tmp/destination",
	)

	var found *models.SyncPair
	found, err = Get(pair.ID)

	assert.NoError(t, err)
	assert.NotNil(t, found)

	assert.Equal(t, pair.ID, found.ID)
}

func TestDeletePair(t *testing.T) {
	root := t.TempDir()
	oldwd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	_ = os.Remove(ConfigFile)

	pair, _ := Add(
		"/tmp/source",
		"/tmp/destination",
	)

	err = Delete(pair.ID)

	assert.NoError(t, err)

	var found *models.SyncPair
	found, err = Get(pair.ID)

	assert.NoError(t, err)
	assert.Nil(t, found)
}
