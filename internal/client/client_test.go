package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	serverapp "FileSync/internal/server"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientSyncFolder(t *testing.T) {
	root := t.TempDir()
	oldwd, err := os.Getwd()
	require.NoError(t, err)
	require.NoError(t, os.Chdir(root))
	t.Cleanup(func() { _ = os.Chdir(oldwd) })

	localRoot := filepath.Join(root, "docs")
	require.NoError(t, os.MkdirAll(localRoot, 0755))
	require.NoError(t, os.WriteFile(filepath.Join(localRoot, "a.txt"), []byte("hello"), 0644))

	app, err := serverapp.NewApp()
	require.NoError(t, err)

	srv := httptest.NewServer(app.Handler())
	t.Cleanup(srv.Close)

	registered, err := RegisterFolder("Docs", localRoot, srv.URL)
	require.NoError(t, err)
	require.NotEmpty(t, registered.ID)

	report, err := SyncFolder(registered.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, report.Added)

	_, err = os.Stat(filepath.Join(root, "storage", "client", "cache", registered.ID+".json"))
	require.NoError(t, err)

	_, err = os.Stat(filepath.Join(root, "storage", "client", "config", "config.json"))
	require.NoError(t, err)

	resp, err := http.Get(srv.URL + "/status")
	require.NoError(t, err)
	defer resp.Body.Close()

	var status struct {
		BlobCount int `json:"blob_count"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&status))
	assert.Equal(t, 1, status.BlobCount)

	_, err = os.Stat(filepath.Join(localRoot, "a.txt"))
	require.NoError(t, err)
}
