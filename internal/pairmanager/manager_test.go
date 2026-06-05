package pairmanager

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddPair(t *testing.T) {

	os.Remove(ConfigFile)

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

	os.Remove(ConfigFile)

	pair, _ := Add(
		"/tmp/source",
		"/tmp/destination",
	)

	found, err := Get(pair.ID)

	assert.NoError(t, err)
	assert.NotNil(t, found)

	assert.Equal(t, pair.ID, found.ID)
}

func TestDeletePair(t *testing.T) {

	os.Remove(ConfigFile)

	pair, _ := Add(
		"/tmp/source",
		"/tmp/destination",
	)

	err := Delete(pair.ID)

	assert.NoError(t, err)

	found, err := Get(pair.ID)

	assert.NoError(t, err)
	assert.Nil(t, found)
}
