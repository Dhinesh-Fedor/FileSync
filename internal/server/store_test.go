package server

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	"FileSync/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestStoreSyncPlan(t *testing.T) {
	root := t.TempDir()
	store, err := NewStore(root)
	require.NoError(t, err)

	folder, err := store.RegisterFolder(models.RegisteredFolder{Name: "Docs", DeviceID: "device-1"})
	require.NoError(t, err)

	plan, err := store.Sync(models.SyncRequest{
		DeviceID:     "device-1",
		Folder:       folder,
		KnownVersion: 0,
		Files: []models.FileInfo{{
			RelativePath: "a.txt",
			SHA256:       "hash1",
			Size:         5,
		}},
	})
	require.NoError(t, err)

	assert.Equal(t, []string{"a.txt"}, plan.UploadPaths)
	assert.Equal(t, int64(0), plan.Version)
	assert.Equal(t, "changes pending upload or deletion", plan.Message)

	status := store.Status()
	require.Len(t, status.Folders, 1)
	assert.Equal(t, int64(0), status.Folders[0].Version)

	digest := sha256.Sum256([]byte("hello"))
	hash := hex.EncodeToString(digest[:])
	_, err = store.StoreBlob(hash, bytes.NewReader([]byte("hello")))
	require.NoError(t, err)
	_, err = store.CommitUpload(folder.ID, "device-1", "a.txt", hash, 5, time.Time{})
	require.NoError(t, err)

	status = store.Status()
	require.Len(t, status.Folders, 1)
	assert.Equal(t, int64(1), status.Folders[0].Version)

	_, err = os.Stat(filepath.Join(root, "blobs"))
	require.NoError(t, err)
}

func TestStoreBlobRejectsHashMismatch(t *testing.T) {
	store, err := NewStore(t.TempDir())
	require.NoError(t, err)

	digest := sha256.Sum256([]byte("expected"))
	hash := hex.EncodeToString(digest[:])
	_, err = store.StoreBlob(hash, bytes.NewReader([]byte("corrupt")))
	require.Error(t, err)

	_, err = os.Stat(filepath.Join(store.root, "blobs", hash))
	assert.Error(t, err)
}

func TestStoreDeleteCreatesTombstone(t *testing.T) {
	store, err := NewStore(t.TempDir())
	require.NoError(t, err)
	folder, err := store.RegisterFolder(models.RegisteredFolder{Name: "Docs", DeviceID: "device-1"})
	require.NoError(t, err)

	digest := sha256.Sum256([]byte("hello"))
	hash := hex.EncodeToString(digest[:])
	_, err = store.StoreBlob(hash, bytes.NewReader([]byte("hello")))
	require.NoError(t, err)
	_, err = store.CommitUpload(folder.ID, "device-1", "a.txt", hash, 5, time.Time{})
	require.NoError(t, err)

	plan, err := store.Sync(models.SyncRequest{
		DeviceID:     "device-1",
		Folder:       folder,
		KnownVersion: 1,
	})
	require.NoError(t, err)
	require.Equal(t, []string{"a.txt"}, plan.DeletePaths)
	require.NoError(t, store.CommitDelete(folder.ID, "device-1", "a.txt"))

	stalePlan, err := store.Sync(models.SyncRequest{
		DeviceID:     "device-1",
		Folder:       folder,
		KnownVersion: 1,
		Files: []models.FileInfo{{
			RelativePath: "a.txt",
			SHA256:       hash,
		}},
	})
	require.NoError(t, err)
	require.Len(t, stalePlan.Conflicts, 1)
	assert.Contains(t, stalePlan.Conflicts[0].Reason, "deleted")
}
